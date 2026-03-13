package object

import "time"

type Account struct {
	ID          string `gorm:"size:128;primarykey"`
	Name        string `gorm:"size:128;not null"`
	Email       string `gorm:"size:320;uniqueIndex"`
	Password    string `gorm:"size:256;not null"`
	TOTPSecret  string `gorm:"size:512;not null"`
	AvatarURL   string `gorm:"size:512;not null"`
	Description string `gorm:"size:512;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
