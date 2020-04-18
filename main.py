from google.cloud import storage

import os
import zlib
from zipfile import ZipFile, ZIP_DEFLATED
from shutil import copyfileobj
import tempfile
import hashlib


def collect_zip(event, context):
# Entry point

    client = storage.Client()

    with tempfile.TemporaryDirectory() as tempdir:
        print(f"Temp dir is {tempdir}")

        zipfilename = None
        try:

            # Download the zip
            bucket = client.get_bucket(event['bucket'])
            name = bucket.get_blob(event['name'])
            print(f'File to process: {bucket}{name}')
            with tempfile.NamedTemporaryFile(delete=False) as temp:
                client.download_blob_to_file(name, temp)
                zipfilename = temp.name

            # Extract the contents
            with ZipFile(zipfilename, 'r') as zipfile:
                zipfile.extractall(path=tempdir)

        finally:

            # Clean up the downloaded zip
            if zipfilename:
                os.remove(zipfilename)

        # Process the extracted files
        for rom in os.listdir(tempdir):
            print(f'Processing {rom}')
            path = os.path.join(tempdir, rom)
            with open(path, 'rb') as r:
                fp = fingerprint(r)
            if exists(rom, fp):
                print("   (exists. Moving on)")
            else:
                store(tempdir, rom, fp)


def fingerprint(input):
# Determine size, crc and sha1

    size = 0
    crc = 0
    hash  = hashlib.sha1()

    while True:
        data = input.read(4096)
        if not data:
            break
        size += len(data)
        hash.update(data)
        crc = zlib.crc32(data, crc)

    sha1 = hash.hexdigest()

    return {
        'size': size,
        'crc': crc,
        'sha1': sha1
    }


def directory_name(fingerprint, name):
# Build the path/key for a rom

    return os.path.join(
        "roms", 
        name, 
        str(fingerprint['size']), 
        str(fingerprint['crc']), 
        fingerprint['sha1']
    )


def store(directory, rom, fingerprint):
# Add a rom to storage

    with tempfile.TemporaryDirectory() as tempdir:

        # Zip the file to a temporary location
        zipname = rom + ".zip"
        zippath = os.path.join(tempdir, zipname)
        with ZipFile(zippath, 'w', compression=ZIP_DEFLATED) as zipfile:
            input = os.path.join(directory, rom)
            zipfile.write(input, arcname=rom)

        # Upload the temporary zip to storage
        client = storage.Client()
        bucket = client.bucket('mamecloud-roms')
        path = directory_name(fingerprint, rom)
        key = os.path.join(path, zipname)
        blob = bucket.blob(key)
        print(f"   storing: {key}")
        blob.upload_from_filename(zippath)


def exists(name, fingerprint):
    ### Check whether this rom exists in storage

    path = directory_name(fingerprint, name)
    filename = os.path.join(path, name + '.zip')

    client = storage.Client()
    for _ in client.list_blobs('mamecloud-roms', prefix=filename, max_results=1):
        return True
    return False


def verify(name, fingerprint):
    ### Verify an existing file in storage is correct

    path = directory_name(fingerprint, name)
    filename = os.path.join(path, name + '.zip')

    client = storage.Client()
    exists = False
    match = False
    found = None
    for blob in client.list_blobs('mamecloud-roms', prefix=filename, max_results=1):
        exists = True
        found = blob

    if exists:
        with tempfile.NamedTemporaryFile(delete=False) as f:
            client.download_blob_to_file(found, f)
            zipname = f.name
        with ZipFile(zipname, 'r') as saved:
            zip_entry = saved.getinfo(name)
            with saved.open(zip_entry) as entry:
                saved_fingerprint = fingerprint(entry)
                size_match = fingerprint['size'] == saved_fingerprint['size']
                crc_match = fingerprint['crc'] == saved_fingerprint['crc']
                sha1_match = fingerprint['sha1'] == saved_fingerprint['sha1']
                match = size_match and crc_match and sha1_match
    
    if not exists:
        print(f"{name} does not exist in storage.")
    elif not match:
        print(f"{name} does not match the stored version. Size:{size_match} crc:{crc_match} sha1:{sha1_match}. Source: {fingerprint}. Saved: {saved_fingerprint}")
    
    return found, match
