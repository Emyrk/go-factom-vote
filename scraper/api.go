package scraper

import (
	"encoding/hex"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
)

const level string = "level"
const bolt string = "bolt"

type Fetcher interface {
	FetchDBlockHead() (interfaces.IDirectoryBlock, error)
	//FetchDBlock(hash interfaces.IHash) (interfaces.IDirectoryBlock, error)
	FetchHeadIndexByChainID(chainID interfaces.IHash) (interfaces.IHash, error)
	FetchEBlock(hash interfaces.IHash) (interfaces.IEntryBlock, error)

	FetchEntry(hash string) (interfaces.IEntry, error)
	FetchDBlockByHeight(dBlockHeight uint32) (interfaces.IDirectoryBlock, error)
	FetchABlockByHeight(blockHeight uint32) (interfaces.IAdminBlock, error)
	FetchFBlockByHeight(blockHeight uint32) (interfaces.IFBlock, error)
	FetchECBlockByHeight(blockHeight uint32) (interfaces.IEntryCreditBlock, error)
	FetchECBlock(keymr interfaces.IHash) (interfaces.IEntryCreditBlock, error)
}

var _ Fetcher = (*APIReader)(nil)

func NewDBReader(levelBolt string, path string) *databaseOverlay.Overlay {
	var dbase *hybridDB.HybridDB
	var err error
	if levelBolt == bolt {
		dbase = hybridDB.NewBoltMapHybridDB(nil, path)
	} else {
		dbase, err = hybridDB.NewLevelMapHybridDB(path, false)
		if err != nil {
			panic(err)
		}
	}

	dbo := databaseOverlay.NewOverlay(dbase)
	return dbo
}

type APIReader struct {
	location string
}

func NewAPIReader(loc string) *APIReader {
	a := new(APIReader)
	a.location = loc
	factom.SetFactomdServer(loc)

	return a
}

func (a *APIReader) FetchEntry(hash string) (interfaces.IEntry, error) {
	raw, err := factom.GetRaw(hash)
	if err != nil {
		return nil, err
	}
	return rawBytesToEntry(raw)
}

func (a *APIReader) FetchEBlock(hash interfaces.IHash) (interfaces.IEntryBlock, error) {
	raw, err := factom.GetRaw(hash.String())
	if err != nil {
		return nil, err
	}
	return rawBytesToEblock(raw)
}

func (a *APIReader) FetchDBlockHead() (interfaces.IDirectoryBlock, error) {
	head, err := factom.GetDBlockHead()
	if err != nil {
		return nil, err
	}
	raw, err := factom.GetRaw(head)
	if err != nil {
		return nil, err
	}
	return rawBytesToDblock(raw)
}

func (a *APIReader) FetchDBlockByHeight(height uint32) (interfaces.IDirectoryBlock, error) {
	raw, err := factom.GetBlockByHeightRaw("d", int64(height))
	if err != nil {
		return nil, err
	}

	data, err := hex.DecodeString(raw.RawData)
	if err != nil {
		return nil, err
	}
	return rawBytesToDblock(data)
}

func (a *APIReader) FetchFBlockByHeight(height uint32) (interfaces.IFBlock, error) {
	raw, err := factom.GetBlockByHeightRaw("f", int64(height))
	if err != nil {
		return nil, err
	}

	data, err := hex.DecodeString(raw.RawData)
	if err != nil {
		return nil, err
	}
	return rawBytesToFblock(data)
}

func (a *APIReader) FetchABlockByHeight(height uint32) (interfaces.IAdminBlock, error) {
	raw, err := factom.GetBlockByHeightRaw("a", int64(height))
	if err != nil {
		return nil, err
	}

	data, err := hex.DecodeString(raw.RawData)
	if err != nil {
		return nil, err
	}
	return rawBytesToAblock(data)
}

func (a *APIReader) FetchECBlock(keymr interfaces.IHash) (interfaces.IEntryCreditBlock, error) {
	data, err := factom.GetRaw(keymr.String())
	if err != nil {
		return nil, err
	}

	return rawBytesToECblock(data)
}

func (a *APIReader) FetchECBlockByHeight(height uint32) (interfaces.IEntryCreditBlock, error) {
	raw, err := factom.GetBlockByHeightRaw("ec", int64(height))
	if err != nil {
		return nil, err
	}

	data, err := hex.DecodeString(raw.RawData)
	if err != nil {
		return nil, err
	}
	return rawBytesToECblock(data)
}

func (a *APIReader) FetchHeadIndexByChainID(chainID interfaces.IHash) (interfaces.IHash, error) {
	resp, err := factom.GetChainHead(chainID.String())
	if err != nil {
		return nil, err
	}
	return primitives.HexToHash(resp)
}

func rawBytesToECblock(raw []byte) (interfaces.IEntryCreditBlock, error) {
	ecblock := entryCreditBlock.NewECBlock()
	err := ecblock.UnmarshalBinary(raw)
	if err != nil {
		return nil, err
	}
	return ecblock, nil
}

func rawBytesToAblock(raw []byte) (interfaces.IAdminBlock, error) {
	ablock := adminBlock.NewAdminBlock(nil)
	err := ablock.UnmarshalBinary(raw)
	if err != nil {
		return nil, err
	}
	return ablock, nil
}

func rawBytesToFblock(raw []byte) (interfaces.IFBlock, error) {
	fblock := factoid.NewFBlock(nil)
	err := fblock.UnmarshalBinary(raw)
	if err != nil {
		return nil, err
	}
	return fblock, nil
}

func rawBytesToDblock(raw []byte) (interfaces.IDirectoryBlock, error) {
	dblock := directoryBlock.NewDirectoryBlock(nil)
	err := dblock.UnmarshalBinary(raw)
	if err != nil {
		return nil, err
	}
	return dblock, nil
}

func rawBytesToEblock(raw []byte) (interfaces.IEntryBlock, error) {
	eblock := entryBlock.NewEBlock()
	err := eblock.UnmarshalBinary(raw)
	if err != nil {
		return nil, err
	}
	return eblock, nil
}

func rawBytesToEntry(raw []byte) (interfaces.IEntry, error) {
	entry := entryBlock.NewEntry()
	err := entry.UnmarshalBinary(raw)
	if err != nil {
		return nil, err
	}
	return entry, nil
}
