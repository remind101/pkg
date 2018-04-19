package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
)

type SNSCarrierPropagator struct {
	publishInput *sns.PublishInput
}

func SNSCarrier(pubInput *sns.PublishInput) *SNSCarrierPropagator {
	return &SNSCarrierPropagator{
		publishInput: pubInput,
	}
}

func (p *SNSCarrierPropagator) Set(key, val string) {
	prefixedKey := addPrefix(key)
	if p.publishInput.MessageAttributes == nil {
		p.publishInput.MessageAttributes = map[string]*sns.MessageAttributeValue{}
	}
	p.publishInput.MessageAttributes[prefixedKey] = &sns.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: &val,
	}
	return
}

func (p *SNSCarrierPropagator) ForeachKey(handler func(key, val string) error) error {
	for k, v := range p.publishInput.MessageAttributes {
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
