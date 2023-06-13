## Buffer

the buffer package uses a piece-table in Go to manage the contents for a text editor.

### Operations

#### Index

Definition: Index(i): return the character at position i
To retrieve the i-th character, the appropriate entry in a piece table is read.

Example
Given the following buffers and piece table:

Buffer	Content

- Original file	`ipsum sit amet`
- Add file	Lorem deletedtext dolor

Piece table

- Which	Start Index	Length
- Add	0	6
- Original	0	6
- Add	17	6
- Original	6	8

To access the i-th character, the appropriate entry in the piece table is looked up.

For instance, to get the value of Index(15), the 3rd entry of piece table is retrieved. This is because the 3rd entry describes the characters from index 12 to 16 (the first entry describes characters in index 0 to 5, the next one is 6 to 11). The piece table entry instructs the program to look for the characters in the "add file" buffer, starting at index 18 in that buffer. The relative index in that entry is 15-12 = 3, which is added to the start position of the entry in the buffer to obtain index of the letter: 3+18 = 21. The value of Index(15) is the 21st character of the "add file" buffer, which is the character "o".

For the buffers and piece table given above, the following text is shown:
```
Lorem ipsum dolor sit amet
```

### Insert

Inserting characters to the text consists of:

- Appending characters to the "add file" buffer, and
- Updating the entry in piece table (breaking an entry into two or three)

### Delete

Single character deletion can be one of two possible conditions:

- The deletion is at the start or end of a piece entry, in which case the appropriate entry in piece table is modified.
- The deletion is in the middle of a piece entry, in which case the entry is split then one of the successor entries is modified as above.
