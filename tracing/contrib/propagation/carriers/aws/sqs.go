package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const SQSCarrierPrefix = "sqs-carrier-"

type SQSCarrierPropagator struct {
	msg *sqs.Message
}

func SQSCarrier(msg *sqs.Message) *SQSCarrierPropagator {
	return &SQSCarrierPropagator{
		msg: msg,
	}
}

func (p *SQSCarrierPropagator) Set(key, val string) {
	prefixedKey := addPrefix(key)
	if p.msg.MessageAttributes == nil {
		p.msg.MessageAttributes = map[string]*sqs.MessageAttributeValue{}
	}
	p.msg.MessageAttributes[prefixedKey] = &sqs.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: &val,
	}
	return
}

func (p *SQSCarrierPropagator) ForeachKey(handler func(key, val string) error) error {
	for k, v := range p.msg.MessageAttributes {
		if v.DataType != nil {
			switch *v.DataType {
			case "String":
				if hasPrefix(k) && v.StringValue != nil {
					keyWithoutPrefix := removePrefix(k)
					handler(keyWithoutPrefix, *v.StringValue)
				}
			default:
			}
		}
	}
	// When should this return an error?
	return nil
}
