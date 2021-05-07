package mongodb

import (
	"context"
	"github.com/patrick246/shortlink/pkg/persistence"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	conn *Connection
}

type Shortlink struct {
	ID  string `bson:"_id"`
	URL string `bson:"url"`
}

var codeCollection = "codes"

func New(conn *Connection) (persistence.Repository, error) {
	return &Repository{
		conn: conn,
	}, nil
}

func (r *Repository) GetLinkForCode(ctx context.Context, code string) (string, error) {
	query := bson.D{{
		"_id", code,
	}}
	sr := r.conn.Collection(codeCollection).FindOne(ctx, query)

	var entry Shortlink
	err := sr.Decode(&entry)
	if err == mongo.ErrNoDocuments {
		return "", persistence.ErrNotFound
	} else if err != nil {
		return "", err
	}

	return entry.URL, nil
}

func (r *Repository) SetLinkForCode(ctx context.Context, code, url string) error {
	entry := bson.D{{
		"$set", bson.D{{
			"url", url,
		}},
	}}

	filter := bson.D{{
		"_id", code,
	}}

	_, err := r.conn.Collection(codeCollection).UpdateOne(ctx, filter, entry, options.Update().SetUpsert(true))
	return err
}

func (r *Repository) GetEntries(ctx context.Context, page, size int64) ([]persistence.Shortlink, int64, error) {
	res, err := r.conn.Collection(codeCollection).Find(ctx, bson.D{}, options.Find().SetLimit(size).SetSkip(page*size))
	if err != nil {
		return nil, 0, err
	}

	var shortlinks []Shortlink
	err = res.All(ctx, &shortlinks)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.conn.Collection(codeCollection).CountDocuments(ctx, bson.D{})
	generic := mapToGeneric(shortlinks)
	if err != nil {
		return generic, int64(len(shortlinks)), nil
	}

	return generic, total, nil
}

func (r *Repository) DeleteCode(ctx context.Context, code string) error {
	_, err := r.conn.Collection(codeCollection).DeleteOne(ctx, bson.D{{"_id", code}})
	if err == mongo.ErrNoDocuments {
		return nil
	}
	return err
}

func mapToGeneric(in []Shortlink) []persistence.Shortlink {
	out := make([]persistence.Shortlink, 0, len(in))
	for _, s := range in {
		out = append(out, persistence.Shortlink{
			Code: s.ID,
			URL:  s.URL,
		})
	}
	return out
}
