package providerid

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	Prefix     = "proxmox://"
	UUIDFormat = `[a-f\d]{8}-[a-f\d]{4}-[a-f\d]{4}-[a-f\d]{4}-[a-f\d]{12}`
)

type ProviderID interface {
	UUID() string
	fmt.Stringer
}

type providerID struct {
	uuid string
}

func New(uuid string) (ProviderID, error) {
	if uuid == "" {
		return nil, errors.New("uuid is required for provider id")
	}

	// to do: validate uuid

	return &providerID{
		uuid: uuid,
	}, nil
}

func (p *providerID) UUID() string {
	return p.uuid
}

func (p *providerID) String() string {
	// provider ID : proxmox://<bios-uuid>
	return Prefix + p.uuid
}
