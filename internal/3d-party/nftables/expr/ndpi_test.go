package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

/*//TODO:                      Не тесты а Адов *Т*Р*Э*Ш
1) убираем все Parallel
2) заменяем assert -> require
3) Убирием эту дичь BaseProtoConverterI и делаем простые и понятные тесты
4) Не делаем экспортируемых функций если на это нет обоснований
5) Тесты на marshal/unmarshal - в помойку ибо - дерьмо полное
   ожилается что ты создаешь лбъект - маршалишь - анмаршалишь и онвый объект идентичен старому

ДАВАЙ УЖЕ ВКЛЮЧАЙСЯ, Я УЖЕ РЕАЛЬНО УСТАЛ РАЗГРЕБАТЬ ЭТО
*/

type ndpiTestSuite struct {
	suite.Suite
}

func (sui *ndpiTestSuite) SetupTest() {
	NdpiModuleProtocolsFile = "./test-data/ndpi-info.txt"
	NdpiState = ndpiLoadInternal()
	sui.Require().NoError(NdpiState.FailReason)
}

func (sui *ndpiTestSuite) bitmask2Proto(mask *ndpiProtoBitmask, expProtos []string) bool {
	ndpi := Ndpi{
		Protocols: nil,
	}
	ndpi.poplulateProtocols(mask)
	return assert.Equal(sui.T(), expProtos, ndpi.Protocols)
}

type BaseProtoConverterI interface {
	proto2Bitmask(protos []string, expMask *ndpiProtoBitmask) (ndpiProtoBitmask, bool)
	bitmask2Proto(mask *ndpiProtoBitmask, expProtos []string) bool
}

type InvalidProto2Bitmask struct {
	*ndpiTestSuite
}

func (sui *InvalidProto2Bitmask) proto2Bitmask(protos []string, expMask *ndpiProtoBitmask) (ndpiProtoBitmask, bool) {
	ndpi := Ndpi{
		Protocols: protos,
	}
	mask, err := ndpi.protocolsToBitmask()
	sui.Require().Error(err)
	if err == nil {
		return mask, false
	}

	return mask, assert.Equal(sui.T(), *expMask, mask)
}

type ValidProto2Bitmask struct {
	*ndpiTestSuite
}

func (sui *ValidProto2Bitmask) proto2Bitmask(protos []string, expMask *ndpiProtoBitmask) (ndpiProtoBitmask, bool) {
	ndpi := Ndpi{
		Protocols: protos,
	}
	mask, err := ndpi.protocolsToBitmask()
	sui.Require().NoError(err)
	if err != nil {
		return mask, false
	}

	return mask, assert.Equal(sui.T(), *expMask, mask)
}

type protoType int

const (
	validProto protoType = iota
	invalidProto
)

func NewProtoConverter(t protoType, s *ndpiTestSuite) BaseProtoConverterI {
	switch t {
	case validProto:
		return &ValidProto2Bitmask{ndpiTestSuite: s}
	case invalidProto:
		return &InvalidProto2Bitmask{ndpiTestSuite: s}
	default:
		return nil
	}
}

func (sui *ndpiTestSuite) Test_NdpiUnSupportedProtocols() {
	sui.T().Parallel() //<-- TODO: Убираем Parallel
	protocols := []string{"memcached", "signal", "xbox", "modbus", "whatsappcall"}
	expMask := ndpiProtoBitmask{}
	pc := NewProtoConverter(invalidProto, sui)
	mask, ok := pc.proto2Bitmask(protocols, &expMask)
	if !ok {
		return
	}
	pc.bitmask2Proto(&mask, nil)
}

func (sui *ndpiTestSuite) Test_NdpiOneSupportedProtocol() {
	sui.T().Parallel() //<-- TODO: Убираем Parallel
	protocols := []string{"http"}
	expMask := ndpiProtoBitmask{
		fds_bits: [NDPI_NUM_FDS_BITS]ndpiMaskType{
			0: 0x80,
		},
	}
	pc := NewProtoConverter(validProto, sui)
	mask, ok := pc.proto2Bitmask(protocols, &expMask)
	if !ok {
		return
	}
	pc.bitmask2Proto(&mask, protocols)
}

func (sui *ndpiTestSuite) Test_NdpiSubstractionProtocols() {
	sui.T().Parallel() //<-- TODO: Убираем Parallel
	protocols := []string{"http", "-http"}
	expMask := ndpiProtoBitmask{}
	pc := NewProtoConverter(validProto, sui)
	mask, ok := pc.proto2Bitmask(protocols, &expMask)
	if !ok {
		return
	}
	pc.bitmask2Proto(&mask, nil)
}

func (sui *ndpiTestSuite) Test_NdpiDublicatedProtocols() {
	sui.T().Parallel() //<-- TODO: Убираем Parallel
	protocols := []string{"http", "http", "http"}
	expMask := ndpiProtoBitmask{
		fds_bits: [NDPI_NUM_FDS_BITS]ndpiMaskType{
			0: 0x80,
		},
	}
	pc := NewProtoConverter(validProto, sui)
	mask, ok := pc.proto2Bitmask(protocols, &expMask)
	if !ok {
		return
	}
	pc.bitmask2Proto(&mask, protocols[:1])
}

func (sui *ndpiTestSuite) Test_NdpiSeveralSuportedProtocols() {
	sui.T().Parallel() //<-- TODO: Убираем Parallel
	protocols := []string{
		NdpiState.Protocols.numbit2ProtoName[7],
		NdpiState.Protocols.numbit2ProtoName[5],
		NdpiState.Protocols.numbit2ProtoName[2],
		NdpiState.Protocols.numbit2ProtoName[9]}

	expMask := ndpiProtoBitmask{
		fds_bits: [NDPI_NUM_FDS_BITS]ndpiMaskType{
			0: 0x2a4,
		},
	}
	pc := NewProtoConverter(validProto, sui)
	mask, ok := pc.proto2Bitmask(protocols, &expMask)
	if !ok {
		return
	}
	//reordered according to bitmask
	pc.bitmask2Proto(&mask, []string{
		NdpiState.Protocols.numbit2ProtoName[2],
		NdpiState.Protocols.numbit2ProtoName[5],
		NdpiState.Protocols.numbit2ProtoName[7],
		NdpiState.Protocols.numbit2ProtoName[9],
	})
}

func (sui *ndpiTestSuite) Test_NdpiMixedProtocols() {
	sui.T().Parallel() //<-- TODO: Убираем Parallel
	protocols := []string{"http", "signal", "dns"}
	expMask := ndpiProtoBitmask{
		fds_bits: [NDPI_NUM_FDS_BITS]ndpiMaskType{
			0: 0x80,
		},
	}
	pc := NewProtoConverter(invalidProto, sui)
	mask, ok := pc.proto2Bitmask(protocols, &expMask)
	if !ok {
		return
	}

	pc.bitmask2Proto(&mask, protocols[:1])
}

func (sui *ndpiTestSuite) Test_NdpiMarshalHostOneProto() {
	sui.T().Parallel() //<-- TODO: Убираем Parallel
	expFlags := NFT_NDPI_FLAG_HOST
	expHost := "youtube"
	expProto := []string{"http"}
	ndpi, err := NewNdpi(NdpiWithHost(expHost), NdpiWithProtocols(expProto[0]))
	sui.Require().NoError(err)
	_, err = ndpi.marshal(0)
	sui.Require().NoError(err)
	assert.Equal(sui.T(), expFlags, ndpi.Flags)
	assert.Equal(sui.T(), expHost, ndpi.Hostname)
	assert.Equal(sui.T(), expProto, ndpi.Protocols)
}

func (sui *ndpiTestSuite) Test_NdpiMarshalHostProtos() {
	sui.T().Parallel() //<-- TODO: Убираем Parallel
	expFlags := NFT_NDPI_FLAG_HOST
	expHost := "youtube"
	expProto := []string{"http", "dns"}
	ndpi, err := NewNdpi(NdpiWithHost(expHost), NdpiWithProtocols(expProto[0], expProto[1]))
	sui.Require().NoError(err)
	_, err = ndpi.marshal(0)
	sui.Require().NoError(err)
	assert.Equal(sui.T(), expFlags, ndpi.Flags)
	assert.Equal(sui.T(), expHost, ndpi.Hostname)
	assert.Equal(sui.T(), expProto, ndpi.Protocols)
}

func (sui *ndpiTestSuite) Test_NdpiMarshalRegex() {
	sui.T().Parallel() //<-- TODO: Убираем Parallel
	expFlags := (NFT_NDPI_FLAG_HOST | NFT_NDPI_FLAG_RE)
	ndpi, err := NewNdpi(NdpiWithHost("/youtube/"), NdpiWithProtocols("http"))
	sui.Require().NoError(err)
	_, err = ndpi.marshal(0)
	sui.Require().NoError(err)
	assert.Equal(sui.T(), expFlags, ndpi.Flags)
}

func (sui *ndpiTestSuite) Test_NdpiMarshalHost() {
	sui.T().Parallel() //<-- TODO: Убираем Parallel
	expFlags := (NFT_NDPI_FLAG_HOST | NFT_NDPI_FLAG_EMPTY)
	ndpi, err := NewNdpi(NdpiWithHost("youtube"))
	sui.Require().NoError(err)
	_, err = ndpi.marshal(0)
	sui.Require().NoError(err)
	assert.Equal(sui.T(), expFlags, ndpi.Flags)
}

func Test_NDPI(t *testing.T) {
	suite.Run(t, new(ndpiTestSuite))
}
