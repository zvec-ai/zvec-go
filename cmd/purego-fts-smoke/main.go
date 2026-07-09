//go:build purego

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	zvec "github.com/zvec-ai/zvec-go"
)

type sampleDoc struct {
	id      string
	content string
	vector  []float32
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := zvec.Initialize(nil); err != nil {
		return fmt.Errorf("initialize zvec: %w", err)
	}
	defer func() {
		if err := zvec.Shutdown(); err != nil {
			log.Printf("shutdown zvec: %v", err)
		}
	}()

	baseDir, err := os.MkdirTemp("", "zvec-purego-fts-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(baseDir)

	collectionPath := filepath.Join(baseDir, "collection")
	collection, err := createCollection(collectionPath)
	if err != nil {
		return err
	}
	defer collection.Close()

	docs := []sampleDoc{
		{"doc1", "The quick brown fox jumps over the lazy dog", []float32{1.0, 0.0, 0.0, 0.0}},
		{"doc2", "A fast red fox runs through the forest", []float32{0.9, 0.1, 0.0, 0.0}},
		{"doc3", "The lazy cat sleeps on the couch all day", []float32{0.0, 1.0, 0.0, 0.0}},
		{"doc4", "Dogs and cats are popular household pets", []float32{0.0, 0.8, 0.2, 0.0}},
		{"doc5", "The fox and the hound is a classic story", []float32{0.7, 0.2, 0.0, 0.1}},
	}
	if err := insertDocs(collection, docs); err != nil {
		return err
	}
	if err := collection.Flush(); err != nil {
		return fmt.Errorf("flush collection: %w", err)
	}

	fmt.Printf("zvec version: %s\n", zvec.GetVersion())
	fmt.Printf("collection: %s\n", collectionPath)

	ftsResults, err := runFTSQuery(collection, "fox")
	if err != nil {
		return err
	}
	defer zvec.FreeDocs(ftsResults)
	if len(ftsResults) == 0 {
		return fmt.Errorf("FTS query returned no results")
	}
	printResults("FTS match fox", ftsResults)

	hybridResults, err := runHybridQuery(collection)
	if err != nil {
		return err
	}
	defer zvec.FreeDocs(hybridResults)
	if len(hybridResults) == 0 {
		return fmt.Errorf("hybrid query returned no results")
	}
	printResults("Hybrid FTS + vector", hybridResults)

	return nil
}

func createCollection(path string) (*zvec.Collection, error) {
	schema := zvec.NewCollectionSchema("purego_fts_smoke")
	if schema == nil {
		return nil, fmt.Errorf("create collection schema")
	}
	defer schema.Destroy()

	idField := zvec.NewFieldSchema("id", zvec.DataTypeString, false, 0)
	if idField == nil {
		return nil, fmt.Errorf("create id field")
	}
	defer idField.Destroy()
	idIndex, err := zvec.NewInvertIndexParams(true, false)
	if err != nil {
		return nil, fmt.Errorf("create id invert index params: %w", err)
	}
	defer idIndex.Destroy()
	if err := idField.SetIndexParams(idIndex); err != nil {
		return nil, fmt.Errorf("set id index params: %w", err)
	}
	if err := schema.AddField(idField); err != nil {
		return nil, fmt.Errorf("add id field: %w", err)
	}

	contentField := zvec.NewFieldSchema("content", zvec.DataTypeString, false, 0)
	if contentField == nil {
		return nil, fmt.Errorf("create content field")
	}
	defer contentField.Destroy()
	ftsIndex, err := zvec.NewFTSIndexParams("whitespace", []string{"lowercase"}, "")
	if err != nil {
		return nil, fmt.Errorf("create content FTS index params: %w", err)
	}
	defer ftsIndex.Destroy()
	if err := contentField.SetIndexParams(ftsIndex); err != nil {
		return nil, fmt.Errorf("set content FTS index params: %w", err)
	}
	if err := schema.AddField(contentField); err != nil {
		return nil, fmt.Errorf("add content field: %w", err)
	}

	vectorField := zvec.NewFieldSchema("embedding", zvec.DataTypeVectorFP32, false, 4)
	if vectorField == nil {
		return nil, fmt.Errorf("create embedding field")
	}
	defer vectorField.Destroy()
	vectorIndex, err := zvec.NewFlatIndexParams(zvec.MetricTypeIP)
	if err != nil {
		return nil, fmt.Errorf("create embedding flat index params: %w", err)
	}
	defer vectorIndex.Destroy()
	if err := vectorField.SetIndexParams(vectorIndex); err != nil {
		return nil, fmt.Errorf("set embedding index params: %w", err)
	}
	if err := schema.AddField(vectorField); err != nil {
		return nil, fmt.Errorf("add embedding field: %w", err)
	}

	collection, err := zvec.CreateAndOpen(path, schema, nil)
	if err != nil {
		return nil, fmt.Errorf("create collection: %w", err)
	}
	return collection, nil
}

func insertDocs(collection *zvec.Collection, samples []sampleDoc) error {
	docs := make([]*zvec.Doc, 0, len(samples))
	for _, sample := range samples {
		doc := zvec.NewDoc()
		if doc == nil {
			return fmt.Errorf("create doc %s", sample.id)
		}
		defer doc.Destroy()
		doc.SetPK(sample.id)
		if err := doc.AddStringField("id", sample.id); err != nil {
			return fmt.Errorf("add id for %s: %w", sample.id, err)
		}
		if err := doc.AddStringField("content", sample.content); err != nil {
			return fmt.Errorf("add content for %s: %w", sample.id, err)
		}
		if err := doc.AddVectorFP32Field("embedding", sample.vector); err != nil {
			return fmt.Errorf("add vector for %s: %w", sample.id, err)
		}
		docs = append(docs, doc)
	}

	result, err := collection.Insert(docs)
	if err != nil {
		return fmt.Errorf("insert docs: %w", err)
	}
	if result.SuccessCount != uint64(len(samples)) || result.ErrorCount != 0 {
		return fmt.Errorf("insert docs: success=%d error=%d", result.SuccessCount, result.ErrorCount)
	}
	fmt.Printf("inserted: %d docs\n", result.SuccessCount)
	return nil
}

func runFTSQuery(collection *zvec.Collection, term string) ([]*zvec.Doc, error) {
	query := zvec.NewSearchQuery()
	if query == nil {
		return nil, fmt.Errorf("create FTS query")
	}
	defer query.Destroy()
	if err := query.SetFieldName("content"); err != nil {
		return nil, fmt.Errorf("set FTS field: %w", err)
	}
	if err := query.SetTopK(10); err != nil {
		return nil, fmt.Errorf("set FTS topk: %w", err)
	}
	if err := query.SetOutputFields([]string{"id", "content"}); err != nil {
		return nil, fmt.Errorf("set FTS output fields: %w", err)
	}

	fts := zvec.NewFTS()
	if fts == nil {
		return nil, fmt.Errorf("create FTS payload")
	}
	defer fts.Destroy()
	if err := fts.SetMatchString(term); err != nil {
		return nil, fmt.Errorf("set FTS match string: %w", err)
	}
	if err := query.SetFTS(fts); err != nil {
		return nil, fmt.Errorf("attach FTS payload: %w", err)
	}

	results, err := collection.Query(query)
	if err != nil {
		return nil, fmt.Errorf("run FTS query: %w", err)
	}
	return results, nil
}

func runHybridQuery(collection *zvec.Collection) ([]*zvec.Doc, error) {
	query := zvec.NewMultiQuery()
	if query == nil {
		return nil, fmt.Errorf("create multi query")
	}
	defer query.Destroy()
	if err := query.SetTopK(3); err != nil {
		return nil, fmt.Errorf("set hybrid topk: %w", err)
	}
	if err := query.SetOutputFields([]string{"id", "content"}); err != nil {
		return nil, fmt.Errorf("set hybrid output fields: %w", err)
	}
	if err := query.SetRerankRRF(60); err != nil {
		return nil, fmt.Errorf("set hybrid RRF rerank: %w", err)
	}

	vectorSub := zvec.NewSubQuery()
	if vectorSub == nil {
		return nil, fmt.Errorf("create vector sub-query")
	}
	defer vectorSub.Destroy()
	if err := vectorSub.SetFieldName("embedding"); err != nil {
		return nil, fmt.Errorf("set vector sub-query field: %w", err)
	}
	if err := vectorSub.SetNumCandidates(5); err != nil {
		return nil, fmt.Errorf("set vector sub-query candidates: %w", err)
	}
	if err := vectorSub.SetQueryVector([]float32{1.0, 0.0, 0.0, 0.0}); err != nil {
		return nil, fmt.Errorf("set vector sub-query vector: %w", err)
	}
	if err := query.AddSubQuery(vectorSub); err != nil {
		return nil, fmt.Errorf("add vector sub-query: %w", err)
	}

	ftsSub := zvec.NewSubQuery()
	if ftsSub == nil {
		return nil, fmt.Errorf("create FTS sub-query")
	}
	defer ftsSub.Destroy()
	if err := ftsSub.SetFieldName("content"); err != nil {
		return nil, fmt.Errorf("set FTS sub-query field: %w", err)
	}
	if err := ftsSub.SetNumCandidates(5); err != nil {
		return nil, fmt.Errorf("set FTS sub-query candidates: %w", err)
	}
	fts := zvec.NewFTS()
	if fts == nil {
		return nil, fmt.Errorf("create hybrid FTS payload")
	}
	defer fts.Destroy()
	if err := fts.SetMatchString("fox"); err != nil {
		return nil, fmt.Errorf("set hybrid FTS match string: %w", err)
	}
	if err := ftsSub.SetFTS(fts); err != nil {
		return nil, fmt.Errorf("attach FTS sub-query: %w", err)
	}
	if err := query.AddSubQuery(ftsSub); err != nil {
		return nil, fmt.Errorf("add FTS sub-query: %w", err)
	}

	results, err := collection.MultiQuery(query)
	if err != nil {
		return nil, fmt.Errorf("run hybrid query: %w", err)
	}
	return results, nil
}

func printResults(label string, docs []*zvec.Doc) {
	fmt.Printf("%s: %d results\n", label, len(docs))
	for i, doc := range docs {
		content, _ := doc.GetStringField("content")
		fmt.Printf("  %d. pk=%s score=%.4f content=%q\n", i+1, doc.GetPK(), doc.GetScore(), content)
	}
}
