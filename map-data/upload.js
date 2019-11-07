// parse args
const args = process.argv.slice(2);
if (args.length != 2) {
  process.stderr.write('usage: node upload.js DATA_FILE MAPBOX_USERNAME\n');
  process.exit(1);
}
const districtsFile = args[0];
const user = args[1];

var accessToken = process.env.MAPBOX_WRITE_SCOPE_ACCESS_TOKEN;
if (accessToken === undefined) {
  process.stderr.write("Missing env var: MAPBOX_WRITE_SCOPE_ACCESS_TOKEN\n");
  process.exit(1);
}

var fs = require('fs'),
    MapboxClient = require('mapbox'),
    AWS = require('aws-sdk');

// get data file base name
var baseName = districtsFile;
if (baseName.indexOf("/") != -1) {
  const parts = baseName.split("/");
  baseName = parts[parts.length - 1];
}
if (baseName.indexOf(".") != -1) {
  baseName = baseName.split(".")[0];
}

var tileset_id = user + "." + baseName; // max 32 characters (including "-labels" added below), only one period
var tileset_name = "US_Congressional_Districts_" + baseName; // max 64 characters (including "_Labels" added below) no spaces

var client = new MapboxClient(accessToken);

// here's how to upload a file to Mapbox

function upload_tileset(file, id, name) {
  client.createUploadCredentials(function(err, credentials) {
    console.log('staging', file, '>', id, '...');

    // Use aws-sdk to stage the file on Amazon S3
    var s3 = new AWS.S3({
         accessKeyId: credentials.accessKeyId,
         secretAccessKey: credentials.secretAccessKey,
         sessionToken: credentials.sessionToken,
         region: 'us-east-1'
    });
    s3.putObject({
      Bucket: credentials.bucket,
      Key: credentials.key,
      Body: fs.createReadStream(file)
    }, function(err, resp) {
      if (err) throw err;

      // Create a Mapbox upload
      client.createUpload({
         tileset: id,
         url: credentials.url,
         name: name
      }, function(err, upload) {
        if (err) throw err;
        console.log(file, id, name, 'uploaded, check mapbox.com/studio for updates.');
      });

    });
  });
}

// do the upload

upload_tileset(districtsFile, tileset_id, tileset_name)
