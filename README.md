# JSON Binary Transmission Format
![Coverage](https://img.shields.io/badge/Coverage-80.4%25-brightgreen)

GLTF inspired JSON schema for embedding arbitrary binaries.

## Schema

### Standard JSON Serialization

Given we have a basic struct `Person`.

```golang
type Person struct {
    Name string
    Age  int
}
```

Serializing it's data:

```golang
out, _ := jbtf.Marshal(Person{
    Name: "Bob",
    Age:  30,
})
log.Print(string(out))
```

Produces JSON that embeds the struct inside a `data` field.

```jsonc
{
    // Data is always present, and takes on the values being serialized
    "data": {
        "Name": "Bob",
        "Age":  30
    }
}
```

Deserializing is as simple of normal JSON deserialization, with generic support

```golang
bob, _ := jbtf.Unmarshal[Person]([]byte(`{"data":{"Name":"Bob","Age":30}}`))
```

### Serializing Binary Data

Let's say we wanted to add a profile picture to the person struct in such a way that we can still load it from a JSON file.

```golang
type Person struct {
    Name            string
    Age             int
    ProfilePicture  *jbtf.Png // Make sure it's always a pointer
}
```

After serializing, we'll notice a few new fields have been added to our JSON document root: `buffers` and `bufferViews`.

```jsonc
{
    "buffers": [
        {
            // Total size of this buffer
            "byteLength": 1000,

            // URI of the buffer. Can either encode directly as a string or
            // reference a seperate binary file containing buffer data
            "uri": "[embedded binary data]"
        }
    ],
    // Array of views into buffers defined within this file.
    "bufferViews": [
        {
            "buffer": 0,       // Index of the buffer we're refering to.
            "byteLength": 1000 // Size in bytes of our view
        }
    ],
    // Data is always present, and takes on the values being serialized
    "data": {
        "Name": "Bob",
        "Age":  30,

        // $ at the start of the field indicates this field has had it's data
        // serialized and written into a buffer.
        // 
        // The value of this field is the index of the buffer view that 
        // contains this fields data.
        "$ProfilePicture": 0
    }
}
```

## Custom Serializers

The marshaller and unmarshaller look for types that implement the `jbtf.Serializable` interface. Let's say we wanted to encode a Person's grade in binary. It's implementation could look something like:

```golang
type Grade struct {
    value float32
}

func (g *Grade) Deserialize(in io.Reader) (err error) {
    data := make([]byte, 4)
    _, err = io.ReadFull(in, data)
    g.value = math.Float32frombits(binary.LittleEndian.Uint32(data))
    return
}

func (pss Grade) Serialize(out io.Writer) error {
    bytes := make([]byte, 4)
    binary.LittleEndian.PutUint32(bytes, math.Float32bits(g.value))
    _, err := out.Write(bytes)
    return err
}
```

Then using the custom serializer is as simple as:

```golang
type Person struct {
    Name            string
    Age             int
    ProfilePicture  *jbtf.Png
    Grade           *Grade    // Make sure it's always a pointer
}
```

Now when serialized, we see a new buffer view appear.

```jsonc
{
    "buffers": [
        {
            "byteLength": 1004,
            "uri": "[embedded binary data]"
        }
    ],
    "bufferViews": [
        // Our custom grade data
        {
            "buffer": 0,    
            "byteLength": 4
        },
        // Profile picture
        {
            "buffer": 0,
            // A byte offset is required if we're not starting from the 
            // begining of the buffer
            "byteOffset": 4,
            "byteLength": 1000
        }
    ],
    "data": {
        "Age":  30,
        "$Grade": 0,
        "Name": "Bob",
        "$ProfilePicture": 1
    }
}
```