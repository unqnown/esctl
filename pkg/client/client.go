package client

import (
	"context"
	"log"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/unqnown/esctl/internal/app"
	"github.com/unqnown/esctl/pkg/dump"
)

type Client struct {
	*elastic.Client
}

func New(cluster app.Cluster, usr app.User, opts ...elastic.ClientOptionFunc) (*Client, error) {
	apply := append(opts, elastic.SetURL(cluster.Servers...))
	if !usr.Nil {
		apply = append(apply, elastic.SetBasicAuth(usr.Name, usr.Password))
	}
	cli, err := elastic.NewSimpleClient(apply...)
	if err != nil {
		return nil, err
	}
	return &Client{Client: cli}, nil
}

func (cli *Client) Bulk() (*Bulker, error) { return NewBulker(cli.Client) }

type Bulker struct {
	*elastic.BulkProcessor
}

func NewBulker(cli *elastic.Client) (*Bulker, error) {
	processor, err := cli.BulkProcessor().
		After(func(execID int64, req []elastic.BulkableRequest, rsp *elastic.BulkResponse, err error) {
			if err != nil {
				log.Printf("bulk processing: executing [%d]: %v", execID, err)
				return
			}
		}).
		Name("esctl_bulk_processor").
		Workers(20).
		BulkActions(100).
		BulkSize(2000 * 100).
		FlushInterval(time.Second).
		Do(context.Background())
	if err != nil {
		return nil, err
	}
	return &Bulker{BulkProcessor: processor}, nil
}

// Add adds requests to remove given ids.
func (b *Bulker) Rm(index string, ids ...string) {
	for _, id := range ids {
		b.rm(index, id)
	}
}

func (b *Bulker) rm(index string, id string) {
	b.Add(
		elastic.NewBulkDeleteRequest().
			Index(index).
			Id(id),
	)
}

func (b *Bulker) Save(docs ...dump.Doc) {
	for _, doc := range docs {
		b.save(doc)
	}
}

func (b *Bulker) save(doc dump.Doc) {
	b.Add(
		elastic.NewBulkIndexRequest().
			Index(doc.Index).
			Id(doc.ID).
			Doc(doc.Body),
	)
}
