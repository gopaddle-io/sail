package db

import (
	"gopaddle/migrationservice/util/context"
	"log"
	"os"
	"strconv"
	"strings"

	mgo "gopkg.in/mgo.v2"
)

type handle struct {
	O        interface{}
	dbName   string
	sessions []*mgo.Session
	total    int
	use      int
}

var instantiated *handle

func Instance() *handle {
	if instantiated == nil {
		instantiated = new(handle)

		//TODO introduce username and password
		count := 32
		if strCount := context.Instance().Get("db-count"); strCount != "" {
			i, err := strconv.Atoi(strCount)
			if err == nil {
				count = i
			}
		}
		log.Println("Connecting to DB:: ", context.Instance().Get("db-name"))
		instantiated.Connect(context.Instance().GetObject("db-endpoint").([]string), context.Instance().Get("db-port"), context.Instance().Get("db-name"), count)
	}
	return instantiated
}

func (h *handle) Connect(endpoint []string, port, dbName string, size int) {
	log.Printf("Creating Conneciton Pool of mgo://%s/%s(%d) ", endpoint, dbName, size)
	for i := 0; i < size; i++ {
		info := &mgo.DialInfo{
			Addrs:    endpoint,
			Database: context.Instance().Get("user-db"),
			Username: context.Instance().Get("db-user"),
			Password: context.Instance().Get("db-password"),
		}

		session, err := mgo.DialWithInfo(info)
		if err != nil {
			log.Printf("Can't connect to mongo, go error %v\n", err)
			os.Exit(1)
		}
		session.SetSafe(&mgo.Safe{})
		h.addSession(session)
	}
	h.total = len(h.sessions)
	h.dbName = dbName

}

func (h *handle) db() *mgo.Database {
	h.use = (h.use + 1) % h.total
	return h.sessions[h.use].DB(h.dbName)
}

func (h *handle) addSession(session *mgo.Session) {
	h.sessions = append(h.sessions, session)
}

func (h *handle) Info() {
	if info, err := h.db().Session.BuildInfo(); err == nil {
		log.Printf("Created mgo://%s (%s), %d Connections", h.dbName, info.Version, h.total)
	} else {
		log.Printf("Error Connecting Database")
		os.Exit(1)
	}
}

func (h *handle) Create(collection string, row interface{}) error {
	//"Id"<= bson.NewObjectId()
	err := h.db().C(collection).Insert(row)
	if err != nil {
		if strings.Contains(err.Error(), "Closed explicitly") {
			log.Println("DataBase Connection Closed Explicitly")
		}
	}
	return err
}

func (h *handle) ReadOne(collection string, condition interface{}) (interface{}, error) {
	var data interface{}
	err := h.db().C(collection).Find(condition).One(&data)
	if err != nil {
		if strings.Contains(err.Error(), "Closed explicitly") {
			log.Println("DataBase Connection Closed Explicitly")
		}
	}
	return data, err
}

func (h *handle) ReadPageBySort(collection string, query interface{}, field string, page int, size int) ([]interface{}, error) {
	var result []interface{}
	skip := size * (page - 1)
	e := h.db().C(collection).Find(query).Limit(size).Skip(skip).Sort(field).All(&result)
	return result, e
}

func (h *handle) FindAndApply(collection string, query interface{}, update interface{}) (*mgo.ChangeInfo, interface{}, error) {
	var result interface{}
	change := mgo.Change{Update: update, ReturnNew: true}
	info, e := h.db().C(collection).Find(query).Apply(change, &result)
	return info, result, e
}

func (h *handle) ReadAll(collection string, query interface{}) (result []interface{}) {
	err := h.db().C(collection).Find(query).All(&result)
	if err != nil {
		log.Println("err: ", err)
	}
	return result
}

func (h *handle) Upsert(collection string, condition interface{}, data interface{}) error {
	_, err := h.db().C(collection).Upsert(condition, data)
	if err != nil {
		log.Printf("Not Found %s", err)
		if strings.Contains(err.Error(), "Closed explicitly") {
			log.Println("DataBase Connection Closed Explicitly")
		}
	}
	return err
}

func (h *handle) Update(collection string, condition interface{}, data interface{}) error {
	err := h.db().C(collection).Update(condition, data)
	if err != nil {
		if strings.Contains(err.Error(), "Closed explicitly") {
			log.Println("DataBase Connection Closed Explicitly")
		}
	}
	return err
}

func (h *handle) UpdateAll(collection string, condition interface{}, data interface{}) {
	var set = make(map[string]interface{})
	set["$set"] = data
	info, err := h.db().C(collection).UpdateAll(condition, set)
	if err != nil {
		log.Printf("Not Found %s", err)
		if strings.Contains(err.Error(), "Closed explicitly") {
			log.Println("DataBase Connection Closed Explicitly")
		}
	}
	log.Println("UpdateAll.Info:", info)
}

func (h *handle) ReadPage(collection string, query interface{}, page int, size int) (result []interface{}) {
	log.Printf("Page:%d, Size:%d/n", page, size)
	skip := size * (page - 1)
	log.Println("Skip_size:", skip)
	h.db().C(collection).Find(query).Limit(size).Skip(skip).All(&result)
	return result
}

func (h *handle) Distinct(collection string, condition interface{}, field string) []interface{} {
	var result []interface{}
	err := h.db().C(collection).Find(condition).Distinct(field, &result)
	if err != nil {
		log.Printf("Not Found %s", err)
		if strings.Contains(err.Error(), "Closed explicitly") {
			log.Println("DataBase Connection Closed Explicitly")
		}
	}
	return result
}

func (h *handle) RemoveOne(collection string, condition interface{}) error {
	err := h.db().C(collection).Remove(condition)
	if err != nil {
		log.Printf("Not Found: %s ", err.Error())
		if strings.Contains(err.Error(), "Closed explicitly") {
			log.Println("DataBase Connection Closed Explicitly")
		}
	}
	return err
}

func (h *handle) RemoveAll(collection string, condition interface{}) error {
	_, err := h.db().C(collection).RemoveAll(condition)
	if err != nil {
		if strings.Contains(err.Error(), "Closed explicitly") {
			log.Println("DataBase Connection Closed Explicitly")
		}
	}
	return err
}

func (h *handle) Sort(collection string, condition interface{}, field string) []interface{} {
	var result []interface{}
	var err error
	if err = h.db().C(collection).Find(condition).Sort(field).All(&result); err != nil {
		if strings.Contains(err.Error(), "Closed explicitly") {
			log.Println("DataBase Connection Closed Explicitly")
		}
	}
	return result
}

func (h *handle) Count(collection string, condition interface{}) (int, error) {
	count, err := h.db().C(collection).Find(condition).Count()
	if err != nil {
		log.Println("Error in fetch count ", count)
		if strings.Contains(err.Error(), "Closed explicitly") {
			log.Println("DataBase Connection Closed Explicitly")
		}
		return count, err
	}
	return count, nil
}
