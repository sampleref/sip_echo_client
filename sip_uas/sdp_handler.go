package sip_uas

import (
	"fmt"
	"github.com/pion/sdp/v3"
	"github.com/rs/zerolog/log"
	"reflect"
	"strings"
)

var (
	AUDIO_CODEC = "PCMU"
	VIDEO_CODEC = "H264"
)

type SDPSession struct {
	callId string

	currOffer  *sdp.SessionDescription
	currAnswer *sdp.SessionDescription
}

func (sdpSession *SDPSession) MarshalSessionDescription(sessionDec *sdp.SessionDescription) ([]byte, error) {
	answerByte, err := sessionDec.Marshal()
	if err != nil {
		log.Err(err)
		return nil, err
	}
	return answerByte, nil
}

func (sdpSession *SDPSession) isCurrentOffer(offer []byte) bool {
	offerParsed := sdp.SessionDescription{}
	if err := offerParsed.Unmarshal(offer); err != nil {
		log.Err(err)
		return false
	}
	var newPorts []int
	for _, mDesc := range offerParsed.MediaDescriptions {
		newPorts = append(newPorts, mDesc.MediaName.Port.Value)
	}
	var currPorts []int
	for _, mDesc := range sdpSession.currOffer.MediaDescriptions {
		currPorts = append(currPorts, mDesc.MediaName.Port.Value)
	}
	log.Info().Msg("Checking Same Offer call Id " + sdpSession.callId)
	fmt.Println(newPorts)
	fmt.Println(currPorts)
	return reflect.DeepEqual(newPorts, currPorts)
}

func (sdpSession *SDPSession) generateAnswer(offer []byte, unicastAddress string, aPort int, vPort int) ([]byte, error) {
	offerParsed := sdp.SessionDescription{}
	if err := offerParsed.Unmarshal(offer); err != nil {
		log.Err(err)
		return nil, err
	}

	sdpSession.currOffer = &offerParsed
	log.Info().Msg("Media Description for CallId >>>>>>>>>>>>>>>>>>>> : " + sdpSession.callId)
	for _, mDesc := range offerParsed.MediaDescriptions {
		log.Info().Msg("Media Description " + mDesc.MediaName.Media + " " + mDesc.MediaName.Port.String())
		fmt.Println(mDesc.MediaName.Formats)
		fmt.Println(mDesc.MediaName.Protos)
		for _, aAttr := range mDesc.Attributes {
			if strings.Contains(aAttr.Value, AUDIO_CODEC) || strings.Contains(aAttr.Value, VIDEO_CODEC) {
				log.Info().Msg("Media Attribute Key: " + aAttr.Key + " Value: " + aAttr.Value)
			} else {
				log.Debug().Msg("Media Attribute Key: " + aAttr.Key + " Value: " + aAttr.Value)
			}
		}
	}
	log.Info().Msg("Media Description >>>>>>>>>>>>>>>>>>>>> ")

	answer := sdp.SessionDescription{
		Version: 0,
		Origin: sdp.Origin{
			Username:       "-",
			SessionID:      offerParsed.Origin.SessionID,
			SessionVersion: offerParsed.Origin.SessionID + 2,
			NetworkType:    "IN",
			AddressType:    "IP4",
			UnicastAddress: unicastAddress,
		},
		SessionName: "SipClient",
		ConnectionInformation: &sdp.ConnectionInformation{
			NetworkType: "IN",
			AddressType: "IP4",
			Address:     &sdp.Address{Address: unicastAddress},
		},
		TimeDescriptions: []sdp.TimeDescription{
			{
				Timing: sdp.Timing{
					StartTime: 0,
					StopTime:  0,
				},
			},
		},
		MediaDescriptions: []*sdp.MediaDescription{
			{
				MediaName: sdp.MediaName{
					Media:   "audio",
					Port:    sdp.RangedPort{Value: aPort},
					Protos:  []string{"RTP", "AVP"},
					Formats: []string{"0"},
				},
				Attributes: []sdp.Attribute{
					{Key: "rtpmap", Value: "0 PCMU/8000"},
					{Key: "ptime", Value: "20"},
					{Key: "maxptime", Value: "150"},
					{Key: "sendrecv"},
				},
			},
			{
				MediaName: sdp.MediaName{
					Media:   "video",
					Port:    sdp.RangedPort{Value: vPort},
					Protos:  []string{"RTP", "AVP"},
					Formats: []string{"103"},
				},
				Attributes: []sdp.Attribute{
					{Key: "rtpmap", Value: "103 H264/90000"},
					{Key: "fmtp", Value: "103 packetization-mode=1;profile-level-id=42e01f"},
					{Key: "sendrecv"},
				},
			},
		},
	}

	sdpSession.currAnswer = &answer
	answerByte, err := answer.Marshal()
	if err != nil {
		log.Err(err)
		return nil, err
	}
	return answerByte, nil
}
