package peer

import "bufio"
import "fmt"
import "regexp"
import "strconv"
import "io"

func mustAtoi(x string) int {
	result, err := strconv.Atoi(x)
	if err != nil {
		panic(err)
	}
	return result
}

// ParseSchema parses a network schema based on a schema dump file
func ParseSchema(schemafile io.Reader) (*NetworkSchema, error) {
	file := bufio.NewReader(schemafile)
	schema := &NetworkSchema{}
	var totalEnums int
	_, err := fmt.Fscanf(file, "%d\n", &totalEnums)
	if err != nil {
		return schema, err
	}

	enumExp := regexp.MustCompile(`\s*"([a-zA-Z0-9 _]+)"\s*(\d+)\s*`)
	schema.Enums = make([]*NetworkEnumSchema, totalEnums)
	for i := 0; i < totalEnums; i++ {
		line, err := file.ReadString('\n')
		if err != nil {
			return schema, err
		}
		enum := enumExp.FindStringSubmatch(line)
		schema.Enums[i] = &NetworkEnumSchema{
			Name:      enum[1],
			BitSize:   uint8(mustAtoi(enum[2])),
			NetworkID: uint16(i),
		}
	}

	var totalInstances int
	var totalProperties int
	var totalEvents int

	instanceExp := enumExp
	eventExp := enumExp
	propertyExp := regexp.MustCompile(`\s*"([a-zA-Z0-9 _\-\(\)/]+)"\s*(\d+)\s*(\d+)\s*`)
	_, err = fmt.Fscanf(file, "%d %d %d\n", &totalInstances, &totalProperties, &totalEvents)
	if err != nil {
		return schema, err
	}
	schema.Instances = make([]*NetworkInstanceSchema, totalInstances)
	schema.Properties = make([]*NetworkPropertySchema, totalProperties)
	schema.Events = make([]*NetworkEventSchema, totalEvents)
	propertyGlobalIndex := 0
	eventGlobalIndex := 0
	for i := 0; i < totalInstances; i++ {
		line, err := file.ReadString('\n')
		if err != nil {
			return schema, err
		}
		instance := instanceExp.FindStringSubmatch(line)
		thisInstance := &NetworkInstanceSchema{
			Name:      instance[1],
			Unknown:   uint16(mustAtoi(instance[2])),
			NetworkID: uint16(i),
		}

		var countProperties int
		_, err = fmt.Fscanf(file, "%d\n", &countProperties)
		if err != nil {
			return schema, err
		}
		thisInstance.Properties = make([]*NetworkPropertySchema, countProperties)

		for j := 0; j < countProperties; j++ {
			line, err = file.ReadString('\n')
			if err != nil {
				return schema, err
			}

			property := propertyExp.FindStringSubmatch(line)
			thisProperty := &NetworkPropertySchema{
				Name:           property[1],
				Type:           uint8(mustAtoi(property[2])),
				EnumID:         uint16(mustAtoi(property[3])),
				TypeString:     TypeNames[uint8(mustAtoi(property[2]))],
				InstanceSchema: thisInstance,
				NetworkID:      uint16(propertyGlobalIndex),
			}
			thisInstance.Properties[j] = thisProperty
			schema.Properties[propertyGlobalIndex] = thisProperty

			propertyGlobalIndex++
		}

		var countEvents int
		_, err = fmt.Fscanf(file, "%d\n", &countEvents)
		if err != nil {
			return schema, err
		}
		thisInstance.Events = make([]*NetworkEventSchema, countEvents)
		for j := 0; j < countEvents; j++ {
			line, err = file.ReadString('\n')
			if err != nil {
				return schema, err
			}

			event := eventExp.FindStringSubmatch(line)
			thisEvent := &NetworkEventSchema{
				Name:           event[1],
				InstanceSchema: thisInstance,
				NetworkID:      uint16(eventGlobalIndex),
			}
			countArguments := mustAtoi(event[2])
			thisEvent.Arguments = make([]*NetworkArgumentSchema, countArguments)
			for k := 0; k < countArguments; k++ {
				var argType int
				var argUnk int

				_, err = fmt.Fscanf(file, "%d %d\n", &argType, &argUnk)
				if err != nil {
					return schema, err
				}
				thisArgument := &NetworkArgumentSchema{
					Type:       uint8(argType),
					TypeString: TypeNames[uint8(argType)],
					EnumID:     uint16(argUnk),
				}

				thisEvent.Arguments[k] = thisArgument
			}

			thisInstance.Events[j] = thisEvent
			schema.Events[eventGlobalIndex] = thisEvent

			eventGlobalIndex++
		}

		schema.Instances[i] = thisInstance
	}
	return schema, nil
}

// Dump encodes a NetworkSchema to a format that can be parsed by ParseSchema()
func (schema *NetworkSchema) Dump(file io.Writer) error {
	var err error

	totalEnums := len(schema.Enums)
	_, err = file.Write([]byte(fmt.Sprintf("%d\n", totalEnums)))
	if err != nil {
		return err
	}
	for _, enum := range schema.Enums {
		_, err = file.Write([]byte(fmt.Sprintf("%q %d\n", enum.Name, enum.BitSize)))
		if err != nil {
			return err
		}
	}

	totalProperties := len(schema.Properties)
	totalEvents := len(schema.Events)
	_, err = file.Write([]byte(fmt.Sprintf("%d %d %d\n", len(schema.Instances), totalProperties, totalEvents)))
	if err != nil {
		return err
	}

	for _, instance := range schema.Instances {
		_, err = file.Write([]byte(fmt.Sprintf("%q %d\n", instance.Name, instance.Unknown)))
		if err != nil {
			return err
		}
		_, err = file.Write([]byte(fmt.Sprintf("\t%d\n", len(instance.Properties))))
		if err != nil {
			return err
		}
		for _, property := range instance.Properties {
			_, err = file.Write([]byte(fmt.Sprintf("\t%q %d %d\n", property.Name, property.Type, property.EnumID)))
			if err != nil {
				return err
			}
		}

		_, err = file.Write([]byte(fmt.Sprintf("\t%d\n", len(instance.Events))))
		if err != nil {
			return err
		}
		for _, event := range instance.Events {
			_, err = file.Write([]byte(fmt.Sprintf("\t%q %d\n", event.Name, len(event.Arguments))))
			if err != nil {
				return err
			}
			for _, argument := range event.Arguments {
				_, err = file.Write([]byte(fmt.Sprintf("\t\t%d %d\n", argument.Type, argument.EnumID)))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
