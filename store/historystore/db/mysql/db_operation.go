package mysql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tendermint/tendermint/libs/log"
)

type DBOperation struct {
	dbHost  string
	dbName  string
	db      *sql.DB
	logHis  log.Logger
}

func NewDBOperation(dbHost string, dbName string, logHis log.Logger) *DBOperation {
	dbOP := new(DBOperation)
	db, err :=  dbOP.init(dbHost, dbName)
	if err != nil || db == nil {
		logHis.Error("dbOP init failed")
		return nil
	} else {
		dbOP.dbHost  = dbHost
		dbOP.dbName  = dbName
		dbOP.db = db
		dbOP.logHis = logHis
	}

	return dbOP
}

func (dbo *DBOperation) init(dbHost string, dbName string) (*sql.DB, error) {
	connStr := fmt.Sprintf("%s/", dbHost)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		dbo.logHis.Error("sql.Open failed", "err", err)
		return nil, err
	}

	_,err = db.Exec("create database if not exists "+ dbName)
	if err != nil {
		dbo.logHis.Error("create db failed", "err", err, "dbName", dbName)
		return nil, err
	}

	db, err = sql.Open("mysql", connStr + dbName)
	if err != nil {
		dbo.logHis.Error("open db failed", "err", err, "dbName", dbName)
		return nil, err
	}

	return db, nil
}

func (dbo *DBOperation) createTable(sql string) error {
	if dbo.db != nil {
		_, err := dbo.db.Exec(sql)
		return err
	}

	return errors.New("dbo.db nil")
}

//insert, update, delete
func (dbo *DBOperation) set(sql string, args ...interface{}) error {
	if dbo.db == nil {
		return errors.New("dbo.db nil")
	}

	st, err := dbo.db.Prepare(sql)
	if err == nil && st != nil {
		defer st.Close()
		_, err = st.Exec(args...)
		return err
	}

	return err
}

func (dbo *DBOperation) query(sql string, args ...interface{}) (string, error) {
	if dbo.db == nil {
		return "", errors.New("dbo.db nil")
	}

	st, err := dbo.db.Prepare(sql)
	if err != nil {
		return "", err
	}

	if  st != nil {
		defer st.Close()
		rows, err := st.Query(args...)
		if err != nil {
			return "", err
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return "", err
		}
		count := len(columns)
		tableData := make([]map[string]interface{}, 0)
		values    := make([]interface{}, count)
		valuePtrs := make([]interface{}, count)
		for rows.Next() {
			for i := 0; i < count; i++ {
				valuePtrs[i] = &values[i]
			}
			rows.Scan(valuePtrs...)
			entry := make(map[string]interface{})
			for i, col := range columns {
				var v interface{}
				val := values[i]
				b, ok := val.([]byte)
				if ok {
					v = string(b)
				} else {
					v = val
				}
				entry[col] = v
			}
			tableData = append(tableData, entry)
		}
		jsonData, err := json.Marshal(tableData)
		if err != nil {
			return "",  err
		}

		strT := string(jsonData)

		return strT[1:len(strT)-1], nil
	} else {
		return "", errors.New("st nil")
	}
}


