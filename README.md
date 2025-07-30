# Go Cloud Native

This projects follows the book [_Cloud Native Go, 2nd Edition_](https://www.oreilly.com/library/view/cloud-native-go/9781098156411/) by Matthew A. Titmus.

Most of the code is adapted from the book, with additionals improvements and enhancements –hopefully– contributed by me.

## TODO

- Add tests
- Add logs
- Add Postman calls for sake of brevity

## Improvements

As suggested by the book, the code was extended as follows. Some features were skipped as they were not relevant for the purposes of this project.

- **_File Transaction Logger_**:

  - Add a Close method to gracefully close the transaction log file.
  - Encode keys and values in the transaction log to handle multi-line/whitespace.
  - Limit the size of keys and values to prevent disk filling.
  - Use a more compact and efficient encoding for the log (not plain text).
  - _(Skipped)_ Implement log compaction to remove records of deleted values.

- **_Postgres Transaction Logger_**:

  - Implement helper functions to create and verify the transactions table.
  - Move database connection parameters (host, port, user, passwrod, dbname) to configuration (not hardcoded).
  - TODO: Add a close method to clean up open database connections.
  - TODO: Ensure all events in the write buffer are flushed to the databse before shutdown (prevent event loss).
  - TODO: Consider implementing log retention or pruning to prevent unbounded growth of the transactions table.
  - TODO: Improve error handling and reporting, specially for database operations.

- **_TLS_**:
  - _(Skipped)_ Implement TLS for the server.
