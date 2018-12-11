// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package statdb

//CtxKey Used as statdb key
type CtxKey int

const (
	ctxKeyStats CtxKey = iota
)

// Config is a configuration struct that is everything you need to start a
// StatDB responsibility
type Config struct {
	DatabaseURL    string `help:"the database connection string to use" default:"$CONFDIR/stats.db"`
	DatabaseDriver string `help:"the database driver to use" default:"sqlite3"`
}

// // Run implements the provider.Responsibility interface
// func (c Config) Run(ctx context.Context, server *provider.Provider) error {
// 	db, ok := ctx.Value("masterdb").(interface {
// 		Irreparable() irreparable.DB
// 	})
// 	if !ok {
// 		return nil, errs.New("unable to get master db instance")
// 	}

// 	return server.Run(context.WithValue(ctx, ctxKeyStats, db))
// }

// // LoadFromContext loads an existing StatDB from the Provider context
// // stack if one exists.
// func LoadFromContext(ctx context.Context) statdb.DB {
// 	db, ok := ctx.Value("masterdb").(interface {
// 		Statdb() statdb.DB
// 	})
// 	if !ok {
// 		return nil, errs.New("unable to get master db instance")
// 	}
// 	return db.Statdb()
// }
