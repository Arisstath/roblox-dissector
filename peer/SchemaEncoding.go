package peer
import "bufio"
import "fmt"
import "regexp"
import "strconv"

func mustAtoi(x string) (int) {
	result, err := strconv.Atoi(x)
	if err != nil {
		panic(err)
	}
	return result
}

func ParseSchema(ifile io.Reader, efile io.Reader) (StaticSchema, error) {
	enums := bufio.NewReader(efile)
	instances := bufio.NewReader(ifile)
	schema := StaticSchema{}
	var totalEnums int
	_, err := fmt.Fscanf(enums, "%d", &enums)
	if err != nil {
		return schema, err
	}

	enumExp := regexp.MustCompile(`"\s*([a-zA-Z0-9 _]+)"\s*\d+\s*"`)
	schema.Enums = make([]StaticEnumSchema, totalEnums)
	for i := 0; i < totalEnums; i++ {
		line, err := enums.ReadString('\n')
		if err != nil {
			return schema, err
		}
		enum := enumExp.FindStringSubmatch(line)
		schema.Enums[i] = StaticEnumSchema{
			Name: enum[1],
			Unknown: MustAtoi(enum[2]),
		}
	}

	var totalInstances int
	var totalProperties int
	var totalEvents int

	instanceExp := enumExp
	eventExp := enumExp
	propertyExp := regexp.MustCompile(`"\s*([a-zA-Z0-9 _]+)"\s*(\d+)\s*(\d+)\s*"`)
	err = fmt.Scanf(instances, "%d %d %d", &totalInstances, totalProperties, totalEvents)
	if err != nil {
		return err
	}
	schema.Instances = make([]StaticInstanceSchema, totalInstances)
	schema.Properties = make([]StaticPropertySchema, totalProperties)
	schema.Events = make([]StaticEventSchema, totalEvents)
	propertyGlobalIndex := 0
	eventGlobalIndex := 0
	for i := 0; i < totalInstances; i++ {
		line, err := instances.ReadString('\n')
		if err != nil {
			return schema, err
		}
		instance := instanceExp.FindStringSubmatch(line)
		thisInstance := StaticInstanceSchema{
			Name: instance[1],
			Unknown: MustAtoi(instance[2]),
		}

		var countProperties int
		_, err = fmt.Fscanf(instances, "%d", &countProperties)
		if err != nil {
			return schema, err
		}
		thisInstance.Properties = make([]StaticPropertySchema, countProperties)

		for j := 0; j < countProperties; j++ {
			line, err = instances.ReadString('\n')
			if err != nil {
				return schema, err
			}

			property := propertyExp.FindStringSubmatch(line)
			thisProperty := StaticPropertySchema{
				Name: property[1],
				Type: MustAtoi(property[2]),
				Unknown: MustAtoi(property[3]),
				TypeString: TypeNames[MustAtoi(property[2])],
				InstanceSchema: thisInstance,
			}
			thisInstance.Properties[j] = thisProperty
			schema.Properties[propertyGlobalIndex] = thisProperty

			propertyGlobalIndex++
		}

		var countEvents int
		_, err = fmt.Fscanf(instances, "%d", &countEvents)
		if err != nil {
			return schema, err
		}
		thisInstance.Properties = make([]StaticEventSchema, countEvents)
		for j := 0; j < countEvents; j++ {
			line, err = instances.ReadString('\n')
			if err != nil {
				return schema, err
			}

			event := eventExp.FindStringSubmatch(line)
			thisEvent := StaticEventSchema{
				Name: property[1],
				InstanceSchema: thisInstance,
			}
			countArguments := MustAtoi(property[2])
			thisEvent.Arguments = make([]StaticArgumentSchema, countArguments)
			for k := 0; k < countArguments; k++ {
				var argType int
				var argUnk int

				fmt.Fscanf(instances, "%d %d", &argType, &argUnk)
				thisArgument := StaticArgumentSchema{
					Type: argType,
					TypeString: TypeNames[argType],
					Unknown: argUnk,
				}

				thisEvent.Arguments[k] = thisArgument
			}

			thisInstance.Events[j] = thisEvent
			schema.Events[eventGlobalIndex] = thisEvent

			eventGlobalIndex++
		}

		schema.Instances[i] = thisInstance
	}
}

func (schema *StaticSchema) Dump(instances io.Writer, enums io.Writer) error {
	var err error

	totalEnums := len(schema.Enums)
	_, err = enums.Write([]byte(fmt.Sprintf("%d\n", totalEnums)))
	if err != nil {
		return err
	}
	for _, enum := range schema.Enums {
		_, err = enums.Write([]byte(fmt.Sprintf("%q %d\n", enum.Name, enum.BitSize)))
		if err != nil {
			return err
		}
	}

	totalProperties := len(schema.Properties)
	totalEvents := len(schema.Events)
	_, err = instances.Write([]byte(fmt.Sprintf("%d %d %d\n", len(schema.Instances), totalProperties, totalEvents)))
	if err != nil {
		return err
	}

	for _, instance := range schema.Instances {
		_, err = instances.Write([]byte(fmt.Sprintf("%q %d\n", instance.Name, instance.Unknown)))
		if err != nil {
			return err
		}
		_, err = instances.Write([]byte(fmt.Sprintf("\t%d\n", len(instance.Properties))))
		if err != nil {
			return err
		}
		for _, property := range instance.Properties {
			_, err = instances.Write([]byte(fmt.Sprintf("\t%q %d %d\n", property.Name, property.Type, property.Unknown)))
			if err != nil {
				return err
			}
		}

		_, err = instances.Write([]byte(fmt.Sprintf("\t%d\n", len(instance.Events))))
		if err != nil {
			return err
		}
		for _, event := range instance.Events {
			_, err = instances.Write([]byte(fmt.Sprintf("\t%q %d\n", event.Name, len(event.Arguments))))
			if err != nil {
				return err
			}
			for _, argument := range event.Arguments {
				_, err = instances.Write([]byte(fmt.Sprintf("\t\t%d %d\n", argument.Type, argument.Unknown)))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
