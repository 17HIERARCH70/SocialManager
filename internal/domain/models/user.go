package models

type User struct {
	ID       int64  `db:"id" json:"id"`
	Email    string `db:"email" json:"email"`
	PassHash []byte `db:"pass_hash" json:"-"`
}
