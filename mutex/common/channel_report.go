package common

import "strings"

type ChannelReport struct {
	Key     string
	Current *Request
}

type ChannelReports []*ChannelReport

func (c ChannelReports) Len() int           { return len(c) }
func (c ChannelReports) Less(i, j int) bool { return strings.Compare(c[i].Key, c[j].Key) < 0 }
func (c ChannelReports) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
