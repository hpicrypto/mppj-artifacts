package api

import (
	"mppj"
	"mppj/api/pb"
)

const (
	pointLen = 33
	ctLen    = 2 * pointLen
)

func GetEncRowMsg(er mppj.EncRow) (*pb.EncRow, error) {
	CuidBytes, err := er.Cuid.Serialize()
	if err != nil {
		return nil, err
	}
	CvalBytes, err := er.Cval[0].Serialize()
	if err != nil {
		return nil, err
	}
	data := make([]byte, 2*ctLen)
	copy(data[0:ctLen], CuidBytes)
	copy(data[ctLen:2*ctLen], CvalBytes)
	return &pb.EncRow{
		Data: data,
	}, nil
}

func GetEncRowFromMsg(msg *pb.EncRow) (mppj.EncRow, error) {
	cuid, err := mppj.DeserializeCiphertext(msg.Data[:ctLen])
	if err != nil {
		return mppj.EncRow{}, err
	}
	cval, err := mppj.DeserializeCiphertext(msg.Data[ctLen : 2*ctLen])
	if err != nil {
		return mppj.EncRow{}, err
	}
	return mppj.EncRow{
		Cuid: cuid,
		Cval: []*mppj.Ciphertext{cval},
	}, nil
}

func GetEncRowWithHintMsg(er mppj.EncRowWithHint) (*pb.EncRowWithHint, error) {
	cnymBytes, err := er.Cnyme.Serialize()
	if err != nil {
		return nil, err
	}
	cvalKeyBytes, err := er.CValKey.Serialize()
	if err != nil {
		return nil, err
	}
	chintBytes, err := er.CHint.Serialize()
	if err != nil {
		return nil, err
	}
	data := make([]byte, 3*ctLen+len(er.CVal))
	copy(data[0:ctLen], cnymBytes)
	copy(data[ctLen:2*ctLen], cvalKeyBytes)
	copy(data[2*ctLen:3*ctLen], chintBytes)
	copy(data[3*ctLen:], er.CVal)
	return &pb.EncRowWithHint{
		Data: data,
	}, nil
}

func GetEncRowWithHintFromMsg(msg *pb.EncRowWithHint) (mppj.EncRowWithHint, error) {
	cnym, err := mppj.DeserializeCiphertext(msg.Data[:ctLen])
	if err != nil {
		return mppj.EncRowWithHint{}, err
	}
	cvalKey, err := mppj.DeserializeCiphertext(msg.Data[ctLen : 2*ctLen])
	if err != nil {
		return mppj.EncRowWithHint{}, err
	}
	chint, err := mppj.DeserializeCiphertext(msg.Data[2*ctLen : 3*ctLen])
	if err != nil {
		return mppj.EncRowWithHint{}, err
	}
	return mppj.EncRowWithHint{
		Cnyme:   *cnym,
		CVal:    msg.Data[3*ctLen:],
		CValKey: *cvalKey,
		CHint:   *chint,
	}, nil
}
