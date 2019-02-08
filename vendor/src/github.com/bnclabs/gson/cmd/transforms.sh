go build -o gson
go build -tags n1ql -o gsonn1ql

echo "list pointers ..."
./gson -pointers -inpfile ../testdata/typical.json
echo

echo "check collation order in directory ..."
./gson -checkdir ../testdata/collate
echo

echo "sort json objects in file ..."
./gsonn1ql -n1qlsort ../testdata/collate/objects.ref
echo

echo "compute overheads ..."
./gson -overheads
echo

echo "value2json ..."
./gson -value2json -inptxt '{"python":"good","perl":"ugly","php":"bad"}'
echo

echo "json2value ..."
./gson -json2value -inptxt '{"python":"good","perl":"ugly","php":"bad"}'
echo

echo "json2cbor ..."
./gson -json2cbor -inptxt '{"python":"good","perl":"ugly","php":"bad"}'
echo

echo "cbor2json ..."
./gson -quote -cbor2json -inptxt '"\xbffpythondgooddperlduglycphpcbad\xff"'
echo

echo "cbor2collate ..."
./gson -quote -cbor2collate -inptxt '"\xbffpythondgooddperlduglycphpcbad\xff"'
echo

echo "collate2cbor ..."
./gson -quote -collate2cbor -inptxt '"xd>3\x00Zperl\x00\x00Zugly\x00\x00Zphp\x00\x00Zbad\x00\x00Zpython\x00\x00Zgood\x00\x00\x00"'
echo

echo "value2cbor ..."
./gson -value2cbor -inptxt '{"python":"good","perl":"ugly","php":"bad"}'
echo

echo "cbor2value ..."
./gson -quote -cbor2value -inptxt '"\xbffpythondgooddperlduglycphpcbad\xff"'
echo

echo "json2collate ..."
./gson -json2collate -inptxt '{"python":"good","perl":"ugly","php":"bad"}'
echo

echo "collate2json ..."
./gson -quote -collate2json -inptxt '"xd>3\x00Zperl\x00\x00Zugly\x00\x00Zphp\x00\x00Zbad\x00\x00Zpython\x00\x00Zgood\x00\x00\x00"'
echo
