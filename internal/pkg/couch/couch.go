package couch

import (
	"context"
	"fmt"
	"net/url"

	_ "github.com/go-kivik/couchdb/v3"
	"github.com/go-kivik/kivik/v3"
)

const _roomsDB = "rooms"

type CouchManager struct {
	client *kivik.Client
}

func New(addr, user, pass string) (*CouchManager, error) {
	url, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("parsing address: %w", err)
	}

	dsn := fmt.Sprintf("%s://%s:%s@%s", url.Scheme, user, pass, url.Host)

	couch, err := kivik.New("couch", dsn)
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}

	return &CouchManager{
		client: couch,
	}, nil
}

func (cm *CouchManager) GetRooms() ([]string, error) {
	docs, err := cm.client.DB(context.TODO(), _roomsDB).AllDocs(context.TODO())
	if err != nil {
		return nil, err
	}

	rooms := []string{}
	for docs.Next() {
		rooms = append(rooms, docs.ID())
	}

	return rooms, nil
}
