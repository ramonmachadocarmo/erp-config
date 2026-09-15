-- CPF/CNPJ becomes optional: some customers (e.g. quick counter sales) have no document on
-- file. The column stays NOT NULL DEFAULT '' (the Go layer already always writes a string,
-- never a null), but the blanket UNIQUE constraint would otherwise reject the second customer
-- ever registered with a blank document. Replace it with a partial unique index that only
-- enforces uniqueness on an actual, non-blank document.
ALTER TABLE people DROP CONSTRAINT people_document_key;
CREATE UNIQUE INDEX people_document_unique_idx ON people (document) WHERE document <> '';
