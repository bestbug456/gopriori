package gopriori

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

var transaction = [][]string{
	[]string{"apple", "beer", "rice", "肉"},
	[]string{"apple", "beer", "rice"},
	[]string{"apple", "beer"},
	[]string{"apple", "pear"},
	[]string{"milk", "beer", "rice", "肉"},
	[]string{"milk", "beer", "rice"},
	[]string{"milk", "beer"},
	[]string{"milk", "pear"},
}

func TestTrain(t *testing.T) {
	a := Train(transaction, Threshold{})
	if a == nil {
		t.Fatalf("Train result is nil")
	}
	t.Log(a)
	support := a.Support([]string{"apple"})
	if support != 0.5 {
		t.Fatalf("Expected %v but having %v", 0.5, support)
	}
	support = a.Support([]string{"apple", "beer", "rice"})
	if support != 0.25 {
		t.Fatalf("Expected %v but having %v", 0.25, support)
	}

	support = a.Support([]string{"apple", "beer", "rice", "肉"})
	if support != 0.125 {
		t.Fatalf("Expected %v but having %v", 0.125, support)
	}

	confidence := a.Confidence([]string{"apple"}, []string{"beer"})
	if confidence != 0.75 {
		t.Fatalf("Expected %v but having %v", 0.75, confidence)
	}

	lift := a.Lift([]string{"apple"}, []string{"beer"})
	if lift != 0.125 {
		t.Fatalf("Expected %v but having %v", 0.125, lift)
	}

	// mixing the order of the elements should return the same result
	support = a.Support([]string{"肉", "rice", "apple", "beer"})
	if support != 0.125 {
		t.Fatalf("Expected %v but having %v", 0.125, support)
	}

	// delete a mixed element
	a.Delete([]string{"肉", "rice", "apple", "beer"})
	support = a.Support([]string{"肉", "rice", "apple", "beer"})
	if support != -1 {
		t.Fatalf("Expected %v but having %v", -1, support)
	}
}

func TestGenerateID(t *testing.T) {
	a := Train(transaction, Threshold{})
	if a == nil {
		t.Fatalf("Train result is nil")
	}
	id := a.generateID([]string{"apple"})
	if id != "apple" {
		t.Fatalf("Expected '%s' but having '%s'", "apple", id)
	}
}

func TestCombined(t *testing.T) {
	ID := map[string][]int{
		"apple": []int{1 << 0},
		"rice":  []int{1 << 1},
		"milk":  []int{1 << 2},
		"肉":     []int{1 << 3},
	}
	t.Log(ID)
	original := map[string][]int{
		"apple": []int{1 << 0},
		"rice":  []int{1 << 1},
		"milk":  []int{1 << 2},
		"肉":     []int{1 << 3},
	}

	combined := combination(original, ID)
	t.Log(combined)
	if len(original["apple"]) != 1 {
		t.Fatalf("Expected %v but having %v", 1, len(original["apple"]))
	}
	if len(combined) != 12 {
		t.Fatalf("Expected %v but having %v", 12, len(combined))
	}
	if value, exist := combined["apple "+"rice"]; !exist || value[0] != 1 || value[1] != 2 {
		t.Fatalf("Expected %v but having %v (exist %v)", 3, value, exist)
	}
	if value, exist := combined["apple "+"milk"]; !exist || value[0] != 1 || value[1] != 4 {
		t.Fatalf("Expected %v but having %v (exist %v)", 5, value, exist)
	}
	if value, exist := combined["apple "+"肉"]; !exist || value[0] != 1 || value[1] != 8 {
		t.Fatalf("Expected %v but having %v (exist %v)", 9, value, exist)
	}
}

func TestCompression(t *testing.T) {
	idmap, matrix := Compression(transaction)
	if len(idmap) != 6 {
		t.Fatalf("Expected %v but having %v", 6, len(idmap))
	}
	if len(matrix) != len(transaction) {
		t.Fatalf("Expected %v but having %v", len(matrix), len(transaction))
	}
}

// BenchmarkTrain-8   	    3142	    393724 ns/op	  108415 B/op	     775 allocs/op
func BenchmarkTrain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Train(transaction, Threshold{})
	}
}

// BenchmarkCombined-8   	10315212	       110 ns/op	      48 B/op	       1 allocs/op
// BenchmarkCombined-8   	  323857	      3681 ns/op	     862 B/op	      15 allocs/op
// BenchmarkCombined-8   	  195996	      5382 ns/op	    1507 B/op	      27 allocs/op
func BenchmarkCombined(b *testing.B) {
	ID := map[string][]int{
		"apple": []int{1 << 0},
		"rice":  []int{1 << 1},
		"milk":  []int{1 << 2},
		"肉":     []int{1 << 3},
	}
	for i := 0; i < b.N; i++ {
		combination(ID, ID)
	}
}

// BenchmarkCompression-8   	  506895	      2284 ns/op	     640 B/op	      11 allocs/op
// BenchmarkCompression-8   	  464547	      2734 ns/op	     832 B/op	      17 allocs/op
func BenchmarkCompression(b *testing.B) {
	b.StopTimer()
	data, err := ioutil.ReadFile("./dataset.json")
	if err != nil {
		b.Fatalf("%s\n", err.Error())
	}
	var benchmarktransactions [][]string
	err = json.Unmarshal(data, &benchmarktransactions)
	if err != nil {
		b.Fatalf("%s\n", err.Error())
	}
	b.Logf("Having %d transactions\n", len(benchmarktransactions))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Compression(benchmarktransactions)
	}
}

// BenchmarkDataset-8			   1  	66073279162 ns/op  	790942320 B/op   4781045 allocs/op
// BenchmarkDataset-8   	       1	12908840891 ns/op	919057352 B/op	 4820532 allocs/op
func BenchmarkDataset(b *testing.B) {
	b.StopTimer()
	data, err := ioutil.ReadFile("./dataset.json")
	if err != nil {
		b.Fatalf("%s\n", err.Error())
	}
	var benchmarktransactions [][]string
	err = json.Unmarshal(data, &benchmarktransactions)
	if err != nil {
		b.Fatalf("%s\n", err.Error())
	}
	b.Logf("Having %d transactions\n", len(benchmarktransactions))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Train(benchmarktransactions, Threshold{})
	}
}
