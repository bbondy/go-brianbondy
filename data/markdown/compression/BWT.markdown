# Burrows-Wheeler Transform

## Lossless data compression

Burrows-Wheeler transform is not actually a data compression algorithm, but it instead is a data transformation algorithm. The algorithm will transfer your input data to a format that is much easier for other compression algorithms to take advantage of.

It does this by re-arranging the data so that duplicate parts appear in a row. For example you will have many 'T' character's in a row instead of spread all over the place.
