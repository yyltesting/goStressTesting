package db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"time"
)
var (
	WSDb *gorm.DB
)
type MsgInfo struct {
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	SendAt int64 `gorm:"column:send_time" json:"send_time"`
	MsgId         string    `gorm:"column:msgid;primary_key" json:"msgid"`
}
func LocalDb() {
	addr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=%t&loc=%s",
		"haboxadmin",
		"boophaiqueibieYaesadootaeGhuo@habox2dev",
		"127.0.0.1:4000",
		"habox",
		true,
		"Local")
	//正式
	//addr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=%t&loc=%s",
	//	"haboxadmin",
	//	"pheisheu7quookeingi1veene4ahTe3o@habox2prd",
	//	"10.0.104.121:4000",
	//	"habox",
	//	true,
	//	"Local")
	db, err := gorm.Open("mysql", addr)
	if err != nil {
		panic("db connect err: " + err.Error())
	}
	db.DB().SetConnMaxLifetime(10 * time.Second)
	db.DB().SetMaxIdleConns(100)
	db.DB().SetMaxOpenConns(2000)
	db = db.Debug()
	WSDb = db
	fmt.Println("DB连接完成")
}
func (m *MsgInfo) TableName() string {
	return "msg_info"
}