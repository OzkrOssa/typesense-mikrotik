package repository

import (
	"gopkg.in/routeros.v2"
)

type MikrotikRepository struct {
	client routeros.Client
}

type Mikrotik interface {
	GetIdentity() (map[string]string, error)
	GetSecrets(btsName string) ([]map[string]string, error)
}

func NewMikrotikRepository(addr, user, password string) (Mikrotik, error) {

	dial, err := routeros.Dial(addr+":8728", user, password)

	if err != nil {
		return nil, err
	}

	return &MikrotikRepository{client: *dial}, nil
}

func (mr *MikrotikRepository) GetIdentity() (map[string]string, error) {
	identity := []map[string]string{}
	mkt, err := mr.client.Run("/system/identity/print")
	if err != nil {
		return nil, err
	}

	for _, r := range mkt.Re {
		identity = append(identity, r.Map)
	}
	return identity[0], nil
}

func (mr *MikrotikRepository) GetSecrets(btsName string) ([]map[string]string, error) {

	secret := []map[string]string{}

	mkt, err := mr.client.Run("/ppp/secret/print")
	if err != nil {
		return nil, err
	}

	for _, d := range mkt.Re {
		d.Map["bts"] = btsName
		secret = append(secret, d.Map)
	}
	return secret, nil
}
