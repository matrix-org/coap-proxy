go build -o gson

echo "value2json ..."
./gson -repeat 100 -mprof value2json.mprof -value2json -inpfile ../../testdata/code.json.gz
go tool pprof --svg --inuse_space gson value2json.mprof > value2json.inuse.svg
go tool pprof --svg --alloc_space gson value2json.mprof > value2json.alloc.svg

echo "json2value ..."
./gson -repeat 100 -mprof json2value.mprof -json2value -inpfile ../../testdata/code.json.gz
go tool pprof --svg --inuse_space gson json2value.mprof > json2value.inuse.svg
go tool pprof --svg --alloc_space gson json2value.mprof > json2value.alloc.svg

echo "value2cbor ..."
./gson -repeat 100 -mprof value2cbor.mprof -value2cbor -inpfile ../../testdata/code.json.gz
go tool pprof --svg --inuse_space gson value2cbor.mprof > value2cbor.inuse.svg
go tool pprof --svg --alloc_space gson value2cbor.mprof > value2cbor.alloc.svg

echo "cbor2value ..."
./gson -repeat 100 -mprof cbor2value.mprof -cbor2value -inpfile ../../testdata/code.cbor.gz
go tool pprof --svg --inuse_space gson cbor2value.mprof > cbor2value.inuse.svg
go tool pprof --svg --alloc_space gson cbor2value.mprof > cbor2value.alloc.svg

echo "json2cbor ..."
./gson -repeat 100 -mprof json2cbor.mprof -json2cbor -inpfile ../../testdata/code.json.gz
go tool pprof --svg --inuse_space gson json2cbor.mprof > json2cbor.inuse.svg
go tool pprof --svg --alloc_space gson json2cbor.mprof > json2cbor.alloc.svg

echo "cbor2json ..."
./gson -repeat 100 -mprof cbor2json.mprof -cbor2json -inpfile ../../testdata/code.cbor.gz
go tool pprof --svg --inuse_space gson cbor2json.mprof > cbor2json.inuse.svg
go tool pprof --svg --alloc_space gson cbor2json.mprof > cbor2json.alloc.svg

echo "cbor2collate ..."
./gson -repeat 100 -mprof cbor2collate.mprof -cbor2collate -inpfile ../../testdata/code.cbor.gz
go tool pprof --svg --inuse_space gson cbor2collate.mprof > cbor2collate.inuse.svg
go tool pprof --svg --alloc_space gson cbor2collate.mprof > cbor2collate.alloc.svg

echo "collate2cbor ..."
./gson -repeat 100 -mprof collate2cbor.mprof -collate2cbor -inpfile ../../testdata/code.collate.gz
go tool pprof --svg --inuse_space gson collate2cbor.mprof > collate2cbor.inuse.svg
go tool pprof --svg --alloc_space gson collate2cbor.mprof > collate2cbor.alloc.svg

echo "json2collate ..."
./gson -repeat 100 -mprof json2collate.mprof -json2collate -inpfile ../../testdata/code.json.gz
go tool pprof --svg --inuse_space gson json2collate.mprof > json2collate.inuse.svg
go tool pprof --svg --alloc_space gson json2collate.mprof > json2collate.alloc.svg

echo "collate2json ..."
./gson -repeat 100 -mprof collate2json.mprof -collate2json -inpfile ../../testdata/code.collate.gz
go tool pprof --svg --inuse_space gson collate2json.mprof > collate2json.inuse.svg
go tool pprof --svg --alloc_space gson collate2json.mprof > collate2json.alloc.svg
