package repository

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
	"go.uber.org/zap"
)

const (
	backupStartPath = "backup_start.db"
	backupStopPath  = "backup_stop.db"
)

type BackupType uint8

const (
	BackupStart BackupType = iota
	BackupStop
)

var ErrDBClosed = fmt.Errorf("database is closed")

type logger struct {
	*zap.SugaredLogger
}

func (l *logger) Warningf(s string, i ...interface{}) {
	l.SugaredLogger.Warnf(s, i...)
}

func NewBadger(dir string, log *zap.Logger) (*badger.DB, error) {
	opts := badger.DefaultOptions(dir)
	opts.Logger = &logger{SugaredLogger: log.Sugar()}
	opts.Compression = options.ZSTD
	opts.ZSTDCompressionLevel = 6

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := Backup(db, BackupStart); err != nil {
		log.Error("backup on start failed", zap.Error(err))
	}

	return db, nil
}

func Backup(db *badger.DB, bt BackupType) error {
	dir := db.Opts().Dir
	var backupPath string

	switch bt {
	case BackupStart:
		backupPath = path.Join(dir, backupStartPath)
	case BackupStop:
		backupPath = path.Join(dir, backupStopPath)
	}

	file, err := os.OpenFile(backupPath, os.O_CREATE|os.O_WRONLY, os.FileMode(0600))
	if err != nil {
		return fmt.Errorf("open backup file %s: %w", backupPath, err)
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = db.Backup(file, 0)
	if err != nil {
		return fmt.Errorf("backup: %w", err)
	}

	return nil
}

func Restore(db *badger.DB) error {
	dir := db.Opts().Dir

	backupPathStart := path.Join(dir, backupStartPath)
	backupPathStop := path.Join(dir, backupStopPath)

	startStat, err := os.Stat(backupPathStart)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat file %s: %w", backupPathStart, err)
	}

	stopStat, err := os.Stat(backupPathStop)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat file %s: %w", backupPathStop, err)
	}

	var backupFile string

	switch {
	case stopStat != nil && startStat != nil:
		if stopStat.ModTime().After(startStat.ModTime()) {
			backupFile = backupPathStop
		} else {
			backupFile = backupPathStart
		}

	case stopStat != nil:
		backupFile = backupPathStart

	case startStat != nil:
		backupFile = backupPathStop
	}

	file, err := os.OpenFile(backupFile, os.O_RDONLY, os.FileMode(0600))
	if err != nil {
		return fmt.Errorf("open backup file %s: %w", backupFile, err)
	}

	defer func() {
		_ = file.Close()
	}()

	if err := db.Load(file, 20); err != nil {
		return fmt.Errorf("load backup: %w", err)
	}

	return nil
}
