package main

import (
	"encoding/json"
	"fmt"
)

type OutboundBody struct {
	Key  string
	Key2 int64
}
type Outbound struct {
	MessageType string
	Body        OutboundBody
}

type Inbound struct {
	MessageType string
	Body        json.RawMessage
}

func main() {
	outbound := Outbound{
		MessageType: "FANCY",
		Body: OutboundBody{
			Key:  "A nice string",
			Key2: 12,
		},
	}

	bytes, err := json.Marshal(outbound)

	if err != nil {
		fmt.Println("Error serializing message", err)
		return
	}

	var inboud = Inbound{}

	err = json.Unmarshal(bytes, &inboud)
	if err != nil {
		fmt.Println("Error deserializing message", err)
		return
	}

	fmt.Println(inboud.MessageType)
	fmt.Println(string(inboud.Body))

	bodyBytes, err := inboud.Body.MarshalJSON()
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	body := OutboundBody{}
	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	fmt.Printf("key -> %s\n", body.Key)
	fmt.Printf("key2 -> %d\n", body.Key2)
}
