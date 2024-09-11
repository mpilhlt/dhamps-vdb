-- This creates an approximate index on the embeddings vectors.

-- Without an index, pg_vector will use a sequential scan to find the nearest neighbors.
-- This will yield perfect recall but will be very slow.
-- PGVector supports two approximate indexes: IVFFlat and HNSW.
-- Both IVFFlat and HNSW are approximate indexes that work with heuristics.
-- That means that there could be errors in their search results.
-- However, they are much faster than exact indexes.

-- HNSW is a good choice for high-dimensional vectors and shines in query time
--   and robustness when data is being updated.
-- IVFFlat is a good choice for low-dimensional vectors and shines in
--   index construction time and memory usage.

-- The m parameter is the number of neighbors to consider during search.
-- The ef_construction parameter is the number of neighbors to consider during index construction.
-- The ef_search parameter is the number of neighbors to consider during search.
-- The vector_size parameter is the size of the vectors in the index.
-- The halfvec_cosine_ops parameter is the cosine similarity operation on half vectors.

-- The HNSW parameters we're using should yield a recall of 0.998 with a decent query time.

-- However, you can only create indexes on rows with the same number of dimensions (using expression and partial indexing):

-- CREATE INDEX embeddings_vector ON embeddings USING hnsw (embedding halfvec_cosine_ops) WITH (m = 16, ef_construction = 128);
CREATE INDEX IF NOT EXISTS embeddings_vector_384  ON embeddings USING hnsw ((embedding::halfvec(384))  halfvec_cosine_ops) WITH (m = 24, ef_construction = 200) WHERE (embedding_dim = 384);  -- Cohere embed-multilingual-light-v3.0, embed-english-light-v3.0
CREATE INDEX IF NOT EXISTS embeddings_vector_768  ON embeddings USING hnsw ((embedding::halfvec(768))  halfvec_cosine_ops) WITH (m = 24, ef_construction = 200) WHERE (embedding_dim = 768);  -- BERT base, Cohere embed-multilingual-v2.0, Gemini Embeddings
CREATE INDEX IF NOT EXISTS embeddings_vector_1024 ON embeddings USING hnsw ((embedding::halfvec(1024)) halfvec_cosine_ops) WITH (m = 24, ef_construction = 200) WHERE (embedding_dim = 1024); -- BERT large, SBERT, Cohere embed-multilingual-v3.0, embed-english-v3.0
CREATE INDEX IF NOT EXISTS embeddings_vector_1536 ON embeddings USING hnsw ((embedding::halfvec(1536)) halfvec_cosine_ops) WITH (m = 24, ef_construction = 200) WHERE (embedding_dim = 1536); -- OpenAI text-embedding-ada-002, text-embedding-3-small
CREATE INDEX IF NOT EXISTS embeddings_vector_3072 ON embeddings USING hnsw ((embedding::halfvec(3072)) halfvec_cosine_ops) WITH (m = 24, ef_construction = 200) WHERE (embedding_dim = 3072); -- OpenAI text-embedding-3-large
-- CREATE INDEX IF NOT EXISTS embeddings_vector_4096 ON embeddings USING hnsw ((embedding::halfvec(4096)) halfvec_cosine_ops) WITH (m = 24, ef_construction = 200) WHERE (embedding_dim = 4096); -- Cohere embed-english-v2.0, Llama 3.1, Mistral 7B, OpenAI text-embedding-ada-001

-- You can then use the index to find the nearest neighbors of a vector:
-- SELECT * FROM embeddings WHERE model_id = 123 ORDER BY embedding::vector(768) <-> '[3,1,2,...]' LIMIT 5;

SET hnsw.ef_search = 100;

---- create above / drop below ----

-- This removes the index on embedding vectors.

DROP INDEX IF EXISTS embeddings_vector_384;
DROP INDEX IF EXISTS embeddings_vector_768;
DROP INDEX IF EXISTS embeddings_vector_1024;
DROP INDEX IF EXISTS embeddings_vector_1536;
DROP INDEX IF EXISTS embeddings_vector_3072;
-- DROP INDEX IF EXISTS embeddings_vector_4096;

SET hnsw.ef_search = 40;
