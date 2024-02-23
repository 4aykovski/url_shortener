package database

import "errors"

var (
	ErrCantCreateDatabase  = errors.New("can't create database")
	ErrCantPingDatabase    = errors.New("can't ping database")
	ErrCantPrepareDatabase = errors.New("can't prepare database")
	ErrURLNotFound         = errors.New("url not found")
	ErrUrlExists           = errors.New("url exists")
	ErrCantSavePage        = errors.New("can't save page")
	ErrCantGetUrl          = errors.New("can't get url")
	ErrCantDeleteUrl       = errors.New("can't delete url")
)
