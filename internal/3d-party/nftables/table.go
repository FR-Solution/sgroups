// Copyright 2018 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nftables

import (
	"encoding/binary"
	"fmt"

	"github.com/google/nftables/binaryutil"
	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

var tableHeaderType = netlink.HeaderType((unix.NFNL_SUBSYS_NFTABLES << 8) | unix.NFT_MSG_NEWTABLE)

// TableFamily specifies the address family for this table.
type TableFamily byte

// Possible TableFamily values.
const (
	TableFamilyUnspecified TableFamily = unix.NFPROTO_UNSPEC
	TableFamilyINet        TableFamily = unix.NFPROTO_INET
	TableFamilyIPv4        TableFamily = unix.NFPROTO_IPV4
	TableFamilyIPv6        TableFamily = unix.NFPROTO_IPV6
	TableFamilyARP         TableFamily = unix.NFPROTO_ARP
	TableFamilyNetdev      TableFamily = unix.NFPROTO_NETDEV
	TableFamilyBridge      TableFamily = unix.NFPROTO_BRIDGE
)

// A Table contains Chains. See also
// https://wiki.nftables.org/wiki-nftables/index.php/Configuring_tables
type Table struct {
	Name   string // NFTA_TABLE_NAME
	Use    uint32 // NFTA_TABLE_USE (Number of chains in table~)
	Flags  uint32 // NFTA_TABLE_FLAGS
	Family TableFamily
}

func (t *Table) flags2netlinkAttr() netlink.Attribute { //это адов костыль
	flags := []byte{0, 0, 0, 0}
	if t.Flags != 0 {
		binary.BigEndian.PutUint32(flags, t.Flags)
	}
	return netlink.Attribute{Type: unix.NFTA_TABLE_FLAGS, Data: flags}
}

// DelTable deletes a specific table, along with all chains/rules it contains.
func (cc *Conn) DelTable(t *Table) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	data := cc.marshalAttr([]netlink.Attribute{
		{Type: unix.NFTA_TABLE_NAME, Data: []byte(t.Name + "\x00")},
		t.flags2netlinkAttr(),
	})
	//data := cc.marshalAttr([]netlink.Attribute{
	//	{Type: unix.NFTA_TABLE_NAME, Data: []byte(t.Name + "\x00")},
	//	{Type: unix.NFTA_TABLE_FLAGS, Data: []byte{0, 0, 0, 0}},
	//})
	cc.messages = append(cc.messages, netlink.Message{
		Header: netlink.Header{
			Type:  netlink.HeaderType((unix.NFNL_SUBSYS_NFTABLES << 8) | unix.NFT_MSG_DELTABLE),
			Flags: netlink.Request | netlink.Acknowledge,
		},
		Data: append(extraHeader(uint8(t.Family), 0), data...),
	})
}

func (cc *Conn) addTable(t *Table, flag netlink.HeaderFlags) *Table {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	data := cc.marshalAttr([]netlink.Attribute{
		{Type: unix.NFTA_TABLE_NAME, Data: []byte(t.Name + "\x00")},
		t.flags2netlinkAttr(),
	})
	//data := cc.marshalAttr([]netlink.Attribute{
	//	{Type: unix.NFTA_TABLE_NAME, Data: []byte(t.Name + "\x00")},
	//	{Type: unix.NFTA_TABLE_FLAGS, Data: []byte{0, 0, 0, 0}},
	//})
	cc.messages = append(cc.messages, netlink.Message{
		Header: netlink.Header{
			Type:  netlink.HeaderType((unix.NFNL_SUBSYS_NFTABLES << 8) | unix.NFT_MSG_NEWTABLE),
			Flags: netlink.Request | netlink.Acknowledge | flag,
		},
		Data: append(extraHeader(uint8(t.Family), 0), data...),
	})
	return t
}

// AddTable adds the specified Table, just like `nft add table ...`.
// See also https://wiki.nftables.org/wiki-nftables/index.php/Configuring_tables
func (cc *Conn) AddTable(t *Table) *Table {
	return cc.addTable(t, netlink.Create)
}

// CreateTable create the specified Table if it do not existed.
// just like `nft create table ...`.
func (cc *Conn) CreateTable(t *Table) *Table {
	return cc.addTable(t, netlink.Excl)
}

// FlushTable removes all rules in all chains within the specified Table. See also
// https://wiki.nftables.org/wiki-nftables/index.php/Configuring_tables#Flushing_tables
func (cc *Conn) FlushTable(t *Table) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	data := cc.marshalAttr([]netlink.Attribute{
		{Type: unix.NFTA_RULE_TABLE, Data: []byte(t.Name + "\x00")},
	})
	cc.messages = append(cc.messages, netlink.Message{
		Header: netlink.Header{
			Type:  netlink.HeaderType((unix.NFNL_SUBSYS_NFTABLES << 8) | unix.NFT_MSG_DELRULE),
			Flags: netlink.Request | netlink.Acknowledge,
		},
		Data: append(extraHeader(uint8(t.Family), 0), data...),
	})
}

// ListTables returns currently configured tables in the kernel
func (cc *Conn) ListTables() ([]*Table, error) {
	return cc.ListTablesOfFamily(TableFamilyUnspecified)
}

// ListTablesOfFamily returns currently configured tables for the specified table family
// in the kernel. It lists all tables if family is TableFamilyUnspecified.
func (cc *Conn) ListTablesOfFamily(family TableFamily) ([]*Table, error) {
	conn, closer, err := cc.netlinkConn()
	if err != nil {
		return nil, err
	}
	defer func() { _ = closer() }()

	msg := netlink.Message{
		Header: netlink.Header{
			Type:  netlink.HeaderType((unix.NFNL_SUBSYS_NFTABLES << 8) | unix.NFT_MSG_GETTABLE),
			Flags: netlink.Request | netlink.Dump,
		},
		Data: extraHeader(uint8(family), 0),
	}

	response, err := conn.Execute(msg)
	if err != nil {
		return nil, err
	}

	var tables []*Table
	for _, m := range response {
		t, err := tableFromMsg(m)
		if err != nil {
			return nil, err
		}

		tables = append(tables, t)
	}

	return tables, nil
}

func tableFromMsg(msg netlink.Message) (*Table, error) {
	if got, want := msg.Header.Type, tableHeaderType; got != want {
		return nil, fmt.Errorf("unexpected header type: got %v, want %v", got, want)
	}

	var t Table
	t.Family = TableFamily(msg.Data[0])

	ad, err := netlink.NewAttributeDecoder(msg.Data[4:])
	if err != nil {
		return nil, err
	}

	for ad.Next() {
		switch ad.Type() {
		case unix.NFTA_TABLE_NAME:
			t.Name = ad.String()
		case unix.NFTA_TABLE_USE:
			t.Use = binaryutil.BigEndian.Uint32(ad.Bytes())
		case unix.NFTA_TABLE_FLAGS:
			if t.Flags = ad.Uint32(); t.Flags != 0 { //это адов костыль
				f0 := binaryutil.NativeEndian.Uint32(binaryutil.BigEndian.PutUint32(unix.NFT_TABLE_F_DORMANT))
				if t.Flags&f0 != 0 {
					t.Flags = unix.NFT_TABLE_F_DORMANT
				}
			}
		}
	}

	return &t, nil
}
