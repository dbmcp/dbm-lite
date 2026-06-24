/*
 * @Project: DBM-Lite 轻量级全域数据库管控平台
 * @Version: v0.1.0
 * @Author: DB老王
 * @License: Apache-2.0 OR MulanPSL-2.0
 */
package model

import (
	"time"
)

type SQLFavorite struct {
	FavoriteID  string    `gorm:"column:favorite_id;primaryKey;size:64" json:"favoriteId"`
	UserID      string    `gorm:"column:user_id;size:64;index" json:"userId"`
	Username    string    `gorm:"column:username;size:128" json:"username"`
	Title       string    `gorm:"column:title;size:256" json:"title"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	SqlText     string    `gorm:"column:sql;type:text" json:"sqlText"`
	CreatedAt   time.Time `gorm:"column:created_at;index" json:"createdAt"`
}

func (SQLFavorite) TableName() string { return "sql_favorites" }