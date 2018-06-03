package table

import (
	"fmt"
	"testing"
	"time"
	"unsafe"

	"github.com/slice-d/genzai/app/codec/gjson"
)

func TestNew(t *testing.T) {
	opts := IncludeString | IncludeInt
	fmt.Println(opts & IncludeFloatAsInt)
}

func TestPoint(t *testing.T) {
	fmt.Println(unsafe.Sizeof(time.Time{}))
	fmt.Println(unsafe.Sizeof(StringKey("")))
	fmt.Println(unsafe.Sizeof(""))
	fmt.Println(unsafe.Sizeof(rectItem{}))
}

func TestCaseInsensitiveCompare(t *testing.T) {
	key := StringCIKey("Beta")
	key2 := StringKey("beta")

	fmt.Println(key.LessThan(key2))
	fmt.Println(key2.LessThan(key))
}

func TestComposite2(t *testing.T) {
	tbl := NewTable()

	tbl.CreateIndex(
		"last_name_age",
		"*",
		JSONComposite(
			JSONIndexer("age", IncludeInt|IncludeFloat|IncludeFloatAsInt),
			JSONIndexer("name.last", IncludeString|CaseInsensitive)))

	//tbl.Set(Key2{StringKey("p"), IntKey(1)}, `{"name":{"first":"Tom","last":"Johnson"},"age":38, "location":[-115.567 33.532]}`, 0)
	tbl.Set(StringKey("p:8"), `{"name":{"first":"Tom","last":"Alpha"},"age":38, "location":[-115.567 33.532]}`, 0)
	tbl.Set(StringKey("p:7"), `{"name":{"first":"Tom","last":"beta"},"age":38, "location":[-115.567 33.532]}`, 0)
	tbl.Set(StringKey("p:6"), `{"name":{"first":"Tom","last":"Beta"},"age":38, "location":[-115.567 33.532]}`, 0)
	tbl.Set(StringKey("p:2"), `{"name":{"first":"Janet","last":"Prichard"},"age":47, "location":[-116.671 35.735]}`, 0)
	tbl.Set(StringKey("p:3"), `{"name":{"first":"Carol","last":"Anderson"},"age":52, "location":[-113.902 31.234]}`, 0)

	fmt.Println("Ascend >=")
	tbl.AscendGreaterOrEqual("last_name_age", Key2{FloatKey(38), StringKey("b")}, func(key IndexItem) bool {
		res := gjson.Get(key.Value().Value, "name.last")
		age := gjson.Get(key.Value().Value, "age")
		fmt.Printf("%s -> %s - %s\n", key.Value().Key, age.Raw, res.Raw)
		return true
	})

	fmt.Println()
	fmt.Println("Ascend <")
	tbl.AscendLessThan("last_name_age", Key2{IntKey(38), StringMax}, func(key IndexItem) bool {
		res := gjson.Get(key.Value().Value, "name.last")
		age := gjson.Get(key.Value().Value, "age")
		fmt.Printf("%s -> %s - %s\n", key.Value().Key, age.Raw, res.Raw)
		return true
	})

	fmt.Println()
	fmt.Println("Descend <=")
	tbl.DescendLessOrEqual("last_name_age", Key2{IntKey(52), StringMin}, func(val IndexItem) bool {
		res := gjson.Get(val.Value().Value, "name.last")
		age := gjson.Get(val.Value().Value, "age")
		fmt.Printf("%s -> %s - %s\n", val.Value().Key, age.Raw, res.Raw)
		return true
	})

	fmt.Println()
	fmt.Println("Descend >")
	tbl.DescendGreaterThan("last_name_age", Key2{IntKey(40), StringMin}, func(key IndexItem) bool {
		res := gjson.Get(key.Value().Value, "name.last")
		age := gjson.Get(key.Value().Value, "age")
		fmt.Printf("%s -> %s - %s\n", key.Value().Key, age.Raw, res.Raw)
		return true
	})

	fmt.Println()
	fmt.Println("Ascend Range")
	tbl.AscendRange("last_name_age", Key2{IntKey(38), StringMin}, Key2{IntKey(52), StringMax}, func(key IndexItem) bool {
		res := gjson.Get(key.Value().Value, "name.last")
		age := gjson.Get(key.Value().Value, "age")
		fmt.Printf("%s -> %s - %s\n", key.Value().Key, age.Raw, res.Raw)
		return true
	})

	fmt.Println()
	fmt.Println("Ascend")
	tbl.Ascend("last_name_age", func(key IndexItem) bool {
		//tbl.AscendRange("age", &floatItem{key: 30}, &floatItem{key: 51}, func(key Value) bool {
		res := gjson.Get(key.Value().Value, "name.last")
		age := gjson.Get(key.Value().Value, "age")
		fmt.Printf("%s -> %s - %s\n", key.Value().Key, age.Raw, res.Raw)
		return true
	})
}

func BenchmarkSortedSet_Ascend(b *testing.B) {
	set := NewTable()

	set.CreateIndex(
		"last_name_age",
		"*",
		JSONComposite(
			JSONIndexer("age", IncludeInt|IncludeFloat|IncludeFloatAsInt),
			JSONIndexer("name.last", IncludeString|CaseInsensitive)))

	//set.Set(Key2{StringKey("p"), IntKey(1)}, `{"name":{"first":"Tom","last":"Johnson"},"age":38, "location":[-115.567 33.532]}`, 0)
	set.Set(StringKey("p:8"), `{"name":{"first":"Tom","last":"Alpha"},"age":38, "location":[-115.567 33.532]}`, 0)
	set.Set(StringKey("p:7"), `{"name":{"first":"Tom","last":"beta"},"age":38, "location":[-115.567 33.532]}`, 0)
	set.Set(StringKey("p:6"), `{"name":{"first":"Tom","last":"Beta"},"age":38, "location":[-115.567 33.532]}`, 0)
	set.Set(StringKey("p:2"), `{"name":{"first":"Janet","last":"Prichard"},"age":47, "location":[-116.671 35.735]}`, 0)
	set.Set(StringKey("p:3"), `{"name":{"first":"Carol","last":"Anderson"},"age":52, "location":[-113.902 31.234]}`, 0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Ascend("last_name_age", func(key IndexItem) bool {
			//set.AscendRange("age", &floatItem{key: 30}, &floatItem{key: 51}, func(key Value) bool {
			//res := gjson.Get(key.Value().Value, "name.last")
			//age := gjson.Get(key.Value().Value, "age")
			//fmt.Printf("%s -> %s - %s\n", key.Value().Extract, age.Raw, res.Raw)
			return true
		})
	}
}

func BenchmarkSortedSet_Set(b *testing.B) {
	set := NewTable()

	set.CreateIndex(
		"last_name_age",
		"*",
		JSONComposite(
			JSONIndexer("age", IncludeInt|IncludeFloat|IncludeFloatAsInt),
			JSONIndexer("name.last", IncludeString|CaseInsensitive)))

	//set.CreateIndex("l", "*", StringIndexer())

	//key := StringKey("p:8")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set.Set(StringKey("p:8"), `{"age":38, "name":{"last":"Alpha", "first":"Tom"}, "location":[-115.567 33.532]}`, 0)
	}
}

func TestSpatial(t *testing.T) {
	db := NewTable()

	//db.CreateJSONStringIndex("last_name", "p:*", "name.last")
	//db.CreateJSONIndex("last_name", "p:*", "name.last")
	db.CreateSpatialIndex("fleet", "fleet:*", SpatialIndexer())
	//db.CreateSpatialIndex("loc", "p:*", "age")

	db.Set(StringKey("fleet:0:pos"), "[-115.567 33.532]", 0)
	db.Set(StringKey("fleet:1:pos"), "[-116.671 35.735]", 0)
	db.Set(StringKey("fleet:2:pos"), "[-113.902 31.234]", 0)

	db.Nearby("fleet", "[-113 33]", func(key Rect, val *ValueItem, dist float64) bool {
		fmt.Println(val.Key, val.Value, dist)
		return true
	})
}

func TestIndexer(t *testing.T) {
	db := NewTable()

	//indexer := NewIndexer("name.last", data.String, false, JSONProjector("name.last"))
	//db.CreateIndex("last_name", "p:*", indexer)

	db.CreateSpatialIndex("fleet", "p:*", JSONSpatialIndexer("location"))

	//db.CreateJSONStringIndex("last_name", "p:*", "name.last")
	//db.CreateJSONIndex("last_name", "p:*", "name.last")
	//db.CreateJSONSpatialIndex("fleet", "p:*", "location")
	//db.CreateSpatialIndex("loc", "p:*", "age")

	db.Set(StringKey("p:1"), `{"name":{"first":"Tom","last":"Johnson"},"age":38, "location":[-115.567 33.532]}`, 0)
	db.Set(StringKey("p:2"), `{"name":{"first":"Janet","last":"Prichard"},"age":47, "location":[-116.671 35.735]}`, 0)
	db.Set(StringKey("p:3"), `{"name":{"first":"Carol","last":"Anderson"},"age":52, "location":[-113.902 31.234]}`, 0)

	db.Nearby("fleet", "[-113 33]", func(key Rect, value *ValueItem, dist float64) bool {
		fmt.Println(value.Key, value.Value, dist)
		return true
	})
}

func TestJsonSpatial(t *testing.T) {
	db := NewTable()

	//db.CreateJSONStringIndex("last_name", "p:*", "name.last")
	//db.CreateJSONIndex("last_name", "p:*", "name.last")
	//db.CreateJSONSpatialIndex("fleet", "p:*", "location")
	//db.CreateSpatialIndex("loc", "p:*", "age")

	db.Set(StringKey("p:1"), `("name":("first":"Tom","last":"Johnson"),"age":38, "location":[-115.567 33.532])`, 0)
	db.Set(StringKey("p:2"), `("name":("first":"Janet","last":"Prichard"),"age":47, "location":[-116.671 35.735])`, 0)
	db.Set(StringKey("p:3"), `("name":("first":"Carol","last":"Anderson"),"age":52, "location":[-113.902 31.234])`, 0)

	//db.Nearby("fleet", "[-113 33]", func(key *rectKey, val *Value, dist float64) bool (
	//	fmt.Println(val.K, val.Value, dist)
	//	return true
	//))
}

func TestRect(t *testing.T) {
	str := EncodeFloat64("30")
	fmt.Println(StrAsFloat64(str))
	fmt.Println(unsafe.Sizeof(FloatKey(0)))
	fmt.Println(unsafe.Sizeof(NilKey{}))
}

func TestSecondary(t *testing.T) {
	//key := &FloatItem{}
	//fmt.Println(unsafe.Offsetof(key.K))

	db := NewTable()

	//db.CreateJSONStringIndex("last_name", "p:*", "name.last")
	//db.CreateIndexM("last_name_age", "p:*", JSONString("name.last"), JSONNumber("age"))
	//db.CreateIndexM("last_name_age", "p:*", JSONNumber("age"), JSONString("name.last"))
	//db.CreateJSONIndex("last_name", "p:*", "name.last")

	//db.CreateJSONNumberIndex("age", "p:*", "age")
	db.CreateIndex(
		"last_name",
		"*",
		JSONIndexer("name.last", IncludeString))
	db.CreateIndex(
		"age",
		"*",
		JSONIndexer("age", IncludeInt|IncludeFloat))

	db.Set(StringKey("p:10"), `{"name":{"first":"Don","last":"Johnson"},"age":38}`, 0)
	db.Set(StringKey("p:1"), `{"name":{"first":"Tom","last":"Johnson"},"age":38}`, 0)
	db.Set(StringKey("p:2"), `{"name":{"first":"Janet","last":"Prichard"},"age":47}`, 0)
	db.Set(StringKey("p:3"), `{"name":{"first":"Carol","last":"Anderson"},"age":52}`, 0)
	db.Set(StringKey("p:4"), `{"name":{"first":"Alan","last":"Cooper"},"age":28}`, 0)
	db.Set(StringKey("p:40"), `{"name":{"first":"Alan","last":30},"age":30}`, 0)
	db.Set(StringKey("p:400"), `{"name":{"first":"Alan","last":28},"age":28}`, 0)
	db.Set(StringKey("p:4000"), `{"name":{"first":"Alan","last":29},"age":29}`, 0)
	db.Set(IntKey(4000), `{"name":{"first":"Alan","last":29},"age":29}`, 0)

	//fmt.Println("Order by last name")
	//db.Ascend("last_name", func(key, value string) bool {
	//	fmt.Printf("%s: %s\n", key, value)
	//	return true
	//})
	//
	//fmt.Println("Order by age")
	//db.Descend("age", func(key, value string) bool {
	//	fmt.Printf("%s: %s\n", key, value)
	//	return true
	//})
	fmt.Println("Table Scan")
	db.AscendPrimary(func(item *ValueItem) bool {
		res := gjson.Get(item.Value, "name.last")
		age := gjson.Get(item.Value, "age")
		fmt.Printf("%s %s: %s\n", item.Key, age.Raw, res.Raw)
		return true
	})

	fmt.Println()
	fmt.Println("Order by age range 30-50")
	db.Descend("last_name", func(key IndexItem) bool {
		res := gjson.Get(key.Value().Value, "name.last")
		age := gjson.Get(key.Value().Value, "age")
		fmt.Printf("%s: %s\n", age.Raw, res.Raw)
		return true
	})

	fmt.Println()
	fmt.Println("Order by age range 30-50")
	db.Ascend("age", func(key IndexItem) bool {
		//db.AscendRange("age", &floatItem{key: 30}, &floatItem{key: 51}, func(key Value) bool {
		res := gjson.Get(key.Value().Value, "name.last")
		age := gjson.Get(key.Value().Value, "age")
		fmt.Printf("%s: %s\n", age.Raw, res.Raw)
		return true
	})
	//db.AscendRange("age", FloatKey(30), FloatKey(52), func(key Value) bool {
	//	//db.AscendRange("age", &floatItem{key: 30}, &floatItem{key: 51}, func(key Value) bool {
	//	res := gjson.SliceForKey(key.Value().Value, "name.last")
	//	age := gjson.SliceForKey(key.Value().Value, "age")
	//	fmt.Printf("%s: %s\n", age.Raw, res.Raw)
	//	return true
	//})
}

func TestDesc(t *testing.T) {
	db := NewTable()

	//db.createSecondaryIndex("lastname", "p:*", &jsonKeyFactory{path: "name.last"}, nil)

	//db.Table("p:10", `{"name":{"first":"Don","last":"Johnson"},"age":38}`, 0)
	//db.Table("p:1", `{"name":{"first":"Tom","last":"Johnson"},"age":38}`, 0)
	//db.Table("p:2", `{"name":{"first":"Janet","last":"Prichard"},"age":47}`, 0)
	//db.Table("p:3", `{"name":{"first":"Carol","last":"Anderson"},"age":52}`, 0)
	//db.Table("p:4", `{"name":{"first":"Alan","last":"Cooper"},"age":28}`, 0)
	//
	//db.Table("a:5", `9`, 0)
	//db.Table("a:6", `47`, 0)
	//db.Table("a:10", `47`, 0)
	//db.Delete("a:10")
	//db.Table("a:7", `52`, 0)
	//db.Table("a:8", `28`, 0)
	//db.Table("a:9", `100`, 0)
	//db.Table(StringKey{"a:9"}, `test`, 0)

	//db.Delete("a:9")
	item := db.get(StringKey(""))
	fmt.Println(item)

	db.DropIndex("age2")

	fmt.Println("Order by last name")
	db.Descend("last_name", func(key IndexItem) bool {
		fmt.Printf("%s: %s\n", key, key.Value().Value)
		return true
	})
	fmt.Println("Order by age")
	db.Ascend("age", func(key IndexItem) bool {
		fmt.Printf("%s: %s\n", key, key.Value().Value)
		return true
	})
	fmt.Println("Order by age range 30-50")
	//db.AscendRange("age", `{"age":30}`, `{"age":50}`, func(key Value) bool {
	//	fmt.Printf("%s: %s\n", key, key.Value().Value)
	//	return true
	//})
	//
	//fmt.Println("Order by age")
	//db.AscendRange("age2", `28`, "50", func(key Value) bool {
	//	fmt.Printf("%s: %s\n", key, key.Value().Value)
	//	return true
	//})

	fmt.Println("Order by age")
	db.Ascend("age2", func(key IndexItem) bool {
		fmt.Printf("%s: %s\n", key, key.Value().Value)
		return true
	})
}

func BenchmarkIndexJSON(b *testing.B) {
}
