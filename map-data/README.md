## Intro

This directory contains code that creates MapBox maps depicting the boundaries
of US states and congressional districts as of a particular session of Congress.

## How to Use

To build the tilesets and style file for the mth through nth Congresses, do:

```
make -j FIRST_CONGRESS=m LAST_CONGRESS=n
```

The results will be put in a directory named "output".

To upload these files to MapBox under user USERNAME, first definine env var
`MAPBOX_WRITE_SCOPE_ACCESS_TOKEN` containing a MapBox access token, and then do:

```
make -j upload FIRST_CONGRESS=m LAST_CONGRESS=n MAPBOX_USER=USERNAME
```


## General Process

This directory does not contain the raw geographical data for these boundaries.
Rather, the code begins by downloading this data.

The code then produces three artifacts for a given session of Congress:
1. A tileset containing the boundaries of the congressional districts
2. A tileset containing the boundaries of the states
3. A MapBox style file that composes these tilesets into one map

Finally, the code uploads these tilesets and the style to MapBox.

## Process Details: Making District Tilesets

For the nth session of Congress:
1. Download shape data for districts for nth Congress from `cdmaps.polisci.ucla.edu`
2. Convert shape file into GeoJSON
3. Process the district features, thus:
   1. Clean up the feature's properties:
      1. Parse the "ID" code into separate "stateFips", "congress", and "district" properties
      2. Add "state" property containing the USPS code for the state
      3. Add "group" property containing "boundary" (used in MapBox style)
   2. Filter out invalid districts
   3. Add short and long titles as feature properties
4. Add label features, one for each district feature
5. Convert to tileset, and save it to `output/n-districts.mbtiles`

## Process Details: Making State Tilesets

For the nth session of Congress:
1. Download shape data for states throughout US history from `publications.newberry.org`
2. Convert shape file into GeoJSON
3. Extract the features for the states of the nth Congress
4. Add label features, one for each state feature
5. Convert to tileset, and save it to `output/n-states.mbtiles`

## Code

The whole thing is orchestrated by the Makefiles in this directory. They are
written to support running them in parallel (i.e., you can pass `-j` when you run
them).

The Makefiles use two 3rd-party programs --- `ogr2ogr` and `tippecanoe` --- and
several Go programs whose code is in the `src` directory.