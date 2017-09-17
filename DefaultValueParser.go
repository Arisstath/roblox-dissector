package main
import "github.com/gskartwii/rbxfile/bin"
import "github.com/gskartwii/rbxfile"
import "os"

func ParseDefaultValues(files []string) DefaultValues {
	result := make(DefaultValues)

	for _, file := range files {
		fd, err := os.Open(file)
		if err != nil {
			println("while opening file:", err.Error())
			continue
		}

		model, err := bin.DeserializeModel(fd, nil)
		if err != nil {
			println("while deserializing:", err.Error())
			continue
		}

		for _, instance := range model.Instances {
			thisClass, ok := result[instance.ClassName]
			if !ok {
				thisClass = make(map[string]rbxfile.Value, len(instance.Properties))
				result[instance.ClassName] = thisClass
			}

			for name, value := range instance.Properties {
				thisClass[name] = value
			}
		}
	}
	println("successfully parsed")

	return result
}
