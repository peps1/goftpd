package acl

import (
	"bytes"
	"sync"

	"github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists       = errors.New("user exists")
	ErrUserDoesntExist  = errors.New("user does not exist")
	ErrGroupExists      = errors.New("group exists")
	ErrGroupDoesntExist = errors.New("group does not exist")
)

type AuthenticatorOpts struct {
	DB string `goftpd:"db"`
}

type Authenticator interface {
	// create
	AddUser(string, string) (*User, error)
	AddGroup(string) (*Group, error)

	// get
	GetUser(string) (*User, error)
	GetGroup(string) (*Group, error)

	// save
	SaveUser(*User) error
	SaveGroup(*Group) error

	// delete
	DeleteUser(user string) error
	DeleteGroup(group string) error

	// utilities
	CheckPassword(string, string) bool
	ChangePassword(string, string) error
}

// Entry describes an Authenticator Entry
type Entry interface {
	Key() []byte
}

// BadgerAuthenticator implements an Authenticator using a badge key/value store
type BadgerAuthenticator struct {
	db         *badger.DB
	bufferPool sync.Pool
}

// NewBadgerAuthenticator takes in options and a badger DB and returns a new BadgerAuthenticator
// which implements the Authenticator interface
func NewBadgerAuthenticator(db *badger.DB) *BadgerAuthenticator {
	return &BadgerAuthenticator{
		db: db,
		bufferPool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

func (a *BadgerAuthenticator) encodeAndUpdate(e Entry) error {
	return a.db.Update(func(tx *badger.Txn) error {
		enc := msgpack.GetEncoder()
		defer msgpack.PutEncoder(enc)

		b := a.bufferPool.Get().(*bytes.Buffer)
		b.Reset()
		defer a.bufferPool.Put(b)

		enc.Reset(b)

		if err := enc.Encode(e); err != nil {
			return err
		}

		return tx.Set(e.Key(), b.Bytes())
	})
}

func (a *BadgerAuthenticator) getAndDecode(key []byte, e Entry) error {
	return a.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			dec := msgpack.GetDecoder()
			defer msgpack.PutDecoder(dec)

			dec.ResetBytes(val)

			if err := dec.Decode(e); err != nil {
				return err
			}

			return nil
		})
	})
}

// AddUser creates a user setting the password
func (a *BadgerAuthenticator) AddUser(name, pass string) (*User, error) {
	// check if we have a user by that name
	u, err := a.GetUser(name)
	if err == nil {
		return nil, ErrUserExists
	}

	if err != ErrUserDoesntExist {
		return nil, err
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u = &User{}

	u.Name = name
	u.Password = hashed

	if err := a.encodeAndUpdate(u); err != nil {
		return nil, err
	}

	return u, nil
}

// AddGroup creates a Group
func (a *BadgerAuthenticator) AddGroup(name string) (*Group, error) {
	// check if we have a user by that name
	g, err := a.GetGroup(name)
	if err == nil {
		return nil, ErrGroupExists
	}

	if err != ErrGroupDoesntExist {
		return nil, err
	}

	g = &Group{}

	g.Name = name

	if err := a.encodeAndUpdate(g); err != nil {
		return nil, err
	}

	return g, nil
}

// GetUser attempts to retrieve a User from the store using the name
func (a *BadgerAuthenticator) GetUser(name string) (*User, error) {
	u := User{Name: name}

	if err := a.getAndDecode(u.Key(), &u); err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrUserDoesntExist
		}
		return nil, err
	}

	return &u, nil
}

// GetGroup attempts to retrieve a Group from the store using the name
func (a *BadgerAuthenticator) GetGroup(name string) (*Group, error) {
	return nil, errors.New("stub")
}

// SaveUser overwrites the User in the store
func (a *BadgerAuthenticator) SaveUser(user *User) error {
	return errors.New("stub")
}

// SaveGroup overwrites the Group in the store
func (a *BadgerAuthenticator) SaveGroup(user *Group) error {
	return errors.New("stub")
}

// DeleteUser removes the User from the store.
// TODO: how to handle shadow fs
func (a *BadgerAuthenticator) DeleteUser(name string) error {
	return errors.New("stub")
}

// DeleteGroup removes the Group from the store and removes it from
// any Users.
// TODO: how to handle shadow fs
func (a *BadgerAuthenticator) DeleteGroup(name string) error {
	return errors.New("stub")
}

// CheckPassword checks to see if the password is the correct one for
// the user. Any failure (i.e. user doesn't exist) returns false.
func (a *BadgerAuthenticator) CheckPassword(name, pass string) bool {
	u, err := a.GetUser(name)
	if err != nil {
		return false
	}

	if err := bcrypt.CompareHashAndPassword(u.Password, []byte(pass)); err != nil {
		return false
	}

	return true
}

// ChangePassword changes the password for the User
func (a *BadgerAuthenticator) ChangePassword(user, pass string) error {
	return errors.New("stub")
}
