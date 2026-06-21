package storage

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"os"
	"time"
)

const (
	OpPut byte = iota
	OpDelete
)

type Storage struct {
	db    *os.File
	index map[string]int64
	close func() error
}

func Init() *Storage {
	file, err := os.OpenFile("data.db", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	s := &Storage{file, make(map[string]int64), file.Close}
	start := time.Now()
	err = s.loadIndex()
	if err != nil {
		s.close()
		log.Fatalln("failed to Load index files :", err)
	}
	log.Println("loaded indexs in", time.Since(start))
	return s
}

func (s *Storage) Write(key string, value []byte) error {

	offset, err := s.db.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	err = binary.Write(s.db, binary.LittleEndian, OpPut)
	if err != nil {
		return err
	}

	err = binary.Write(s.db, binary.LittleEndian, uint32(len(key)))
	if err != nil {
		return err
	}
	err = binary.Write(s.db, binary.LittleEndian, uint32(len(value)))
	if err != nil {
		return err
	}

	_, err = s.db.Write([]byte(key))
	if err != nil {
		return err
	}
	_, err = s.db.Write(value)
	if err != nil {
		return err
	}

	s.index[key] = offset
	return nil
}

func (s *Storage) Read(key string) ([]byte, error) {
	offset, ok := s.index[key]
	if !ok {
		return nil, ErrOffsetNotFound
	}

	_, err := s.db.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	var op byte

	err = binary.Read(s.db, binary.LittleEndian, &op)
	if err != nil {
		return nil, err
	}

	if op == OpDelete {
		return nil, ErrOffsetNotFound
	}

	var keyLen, valLen uint32

	err = binary.Read(s.db, binary.LittleEndian, &keyLen)
	if err != nil {
		return nil, err
	}

	err = binary.Read(s.db, binary.LittleEndian, &valLen)
	if err != nil {
		return nil, err
	}

	// skip key
	_, err = s.db.Seek(int64(keyLen), io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	value := make([]byte, valLen)
	err = binary.Read(s.db, binary.LittleEndian, &value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (s *Storage) Delete(key string) error {

	if _, err := s.db.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	err := binary.Write(s.db, binary.LittleEndian, OpDelete)
	if err != nil {
		return err
	}
	err = binary.Write(s.db, binary.LittleEndian, uint32(len(key)))
	if err != nil {
		return err
	}
	err = binary.Write(s.db, binary.LittleEndian, uint32(0))
	if err != nil {
		return err
	}

	_, err = s.db.Write([]byte(key))
	if err != nil {
		return err
	}

	delete(s.index, key)

	return nil
}

func (s *Storage) loadIndex() error {
	if _, err := s.db.Seek(0, io.SeekStart); err != nil {
		return err
	}

	for {
		offset, _ := s.db.Seek(0, io.SeekCurrent)

		var op byte

		if err := binary.Read(s.db, binary.LittleEndian, &op); err != nil {
			if err == io.EOF {
				break
			}
			log.Println("failed to read op ", err)
			return err
		}

		var keyLen, valueLen uint32
		if err := binary.Read(s.db, binary.LittleEndian, &keyLen); err != nil {
			if err == io.EOF {
				log.Println("eof len read :", err)
				break
			}
			log.Println("failed to read key len :", err)
			return err
		}

		if err := binary.Read(s.db, binary.LittleEndian, &valueLen); err != nil {
			log.Println("failed to read value len :", err)
			return err
		}

		key := make([]byte, keyLen)
		if _, err := io.ReadFull(s.db, key); err != nil {
			log.Println("failed to read key :", err)
			return err
		}

		if _, err := s.db.Seek(int64(valueLen), io.SeekCurrent); err != nil {
			log.Println("failed to skip value len :", err)
			return err
		}

		switch op {
		case OpPut:
			s.index[string(key)] = offset
		case OpDelete:
			delete(s.index, string(key))

		}

	}

	return nil
}

var (
	ErrOffsetNotFound = errors.New("offset not build yet.")
)
