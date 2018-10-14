# IGCAPI

## Description

This is an assignment project written in Golang
 that uses the `goigc` library and is deployed on Google Cloud app-engine.
The idea of this project is to extract data from IGC files and
output them as JSON.
The application is deployed on 
here http://igcapifly.appspot.com/igcinfo/api/igc

## Usage

To see the application's metadata navigate to 
`igcinfo/api`

URLs need to be posted at
`igcinfo/api/igc`
and will be stored with a unique id starting from '0'.

To display JSON from a given URL navigate to
`igcinfo/api/igc/<id>`

To only display a field of an IGC file as plain text, navigate to
`igcinfo/api/<id>/<field>`

A field can only be the following:
```
pilot
glider
glider_id
track_length
H_date
```



## Authors
Askel Eirik Johansson

