# Go Cloud Native

This projects follows the book [_Cloud Native Go, 2nd Edition_](https://www.oreilly.com/library/view/cloud-native-go/9781098156411/) by Matthew A. Titmus.

Most of the code is adapted from the book, with additionals improvements and enhancements –hopefully– contributed by me.

## TODO

- Add tests

## Improvements

As suggested by the book, the code was extended as follows. Some features were skipped as they were not relevant for the purposes of this project.

- **_Transaction log_**:

  - Add a Close method to gracefully close the transaction log file.
  - Encode keys and values in the transaction log to handle multi-line/whitespace.
  - Limit the size of keys and values to prevent disk filling.
  - Use a more compact and efficient encoding for the log (not plain text).
  - _(Skipped)_ Implement log compaction to remove records of deleted values.

- **_Postgres_**:
  - Implement helper functions to create and verify the transactions table.
