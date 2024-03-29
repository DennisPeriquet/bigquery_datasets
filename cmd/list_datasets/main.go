package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// TableInfo holds information about a BigQuery table.
// PartitioningType will be calculated as one of: "Time", "Range", or "" (i.e., none)
// based on the type of partitioning the table has.
type TableInfo struct {
	DatasetID        string
	TableID          string
	LastModifiedTime time.Time
	NumRows          uint64
	IsPartitioned    bool
	PartitioningType string
	IsMaterialized   bool
}

func main() {
	ctx := context.Background()

	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <projectID> <keyPath>", os.Args[0])
		fmt.Println()
		fmt.Println("Given a Bigquery projectID, list the tables for all datasets accessible by the given credentials")
		os.Exit(1)
	}

	// Pull the projectID and keyPath from the command line.
	projectID := os.Args[1]
	keyPath := os.Args[2]

	// If the project ID string does not occur in the key path, this is probably
	// the wrong project or key path.
	if !strings.Contains(keyPath, projectID) {
		log.Fatalf("The project ID %s does not appear in the key path %s (and is usually an error)", projectID, keyPath)
	}

	client, err := bigquery.NewClient(ctx, projectID, option.WithCredentialsFile(keyPath))
	if err != nil {
		log.Fatalf("Failed to create bigquery client: %v", err)
	}

	const maxTables = 100
	datasets, err := listTablesByDataset(ctx, client, maxTables)
	if err != nil {
		log.Fatalf("Failed to list tables: %v", err)
	}

	fmt.Printf("Project: %s\n\n", projectID)
	for datasetID, tables := range datasets {
		fmt.Printf("\nDataset: %s, Tables: %d\n", datasetID, len(tables))
		for _, table := range tables {
			partitionedStatus := "Not Partitioned"
			if table.IsPartitioned {
				partitionedStatus = fmt.Sprintf("Partitioned (%s)", table.PartitioningType)
			}
			materializedStatus := ""
			if table.IsMaterialized {
				materializedStatus = ", Materialized"
			}
			fmt.Printf("  %35v, %12d, %s, %s%s\n", table.LastModifiedTime, table.NumRows, table.TableID, partitionedStatus, materializedStatus)

			// Get the list of columns in the table.
			meta, err := client.Dataset(datasetID).Table(table.TableID).Metadata(ctx)
			if err != nil {
				log.Fatalf("Failed to get table metadata: %v", err)
			}
			for _, schema := range meta.Schema {
				fmt.Printf("        %-20s: %s, Description=%s\n", schema.Name, schema.Type, schema.Description)
			}
		}
	}
}

// listTablesByDataset returns a map of datasetID to a slice of TableInfo for all tables
// in the dataset We only read limited number of tables to avoid long running time
// (e.g., for debugging)
func listTablesByDataset(ctx context.Context, client *bigquery.Client, maxTables int) (map[string][]TableInfo, error) {
	datasets := make(map[string][]TableInfo)

	it := client.Datasets(ctx)
	for {
		dataset, err := it.Next()

		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.WithMessage(err, "Failed to get next dataset")
		}
		// Skip the ones with "_test" as those are for individuals experimenting and testing
		if strings.Contains(dataset.DatasetID, "_test") {
			continue
		}

		tables := client.Dataset(dataset.DatasetID).Tables(ctx)
		var datasetTables []TableInfo

		maxTablesForDataset := maxTables
		for {
			table, err := tables.Next()
			if err == iterator.Done {
				break
			}
			if maxTablesForDataset < 0 {
				break
			}
			maxTablesForDataset--
			if err != nil {
				return nil, errors.WithMessage(err, "Failed to get next table")
			}
			meta, err := table.Metadata(ctx)
			if err != nil {
				return nil, errors.WithMessage(err, "Failed to get table metadata")
			}
			tableInfo := TableInfo{
				DatasetID:        dataset.DatasetID,
				TableID:          table.TableID,
				LastModifiedTime: meta.LastModifiedTime,
				NumRows:          meta.NumRows,
				IsPartitioned:    meta.TimePartitioning != nil || meta.RangePartitioning != nil,
				IsMaterialized:   meta.Type == bigquery.MaterializedView,
			}

			// Calculate the partitioning type so we only need one field to display it.
			if meta.TimePartitioning != nil {
				tableInfo.PartitioningType = "Time"
			} else if meta.RangePartitioning != nil {
				tableInfo.PartitioningType = "Range"
			}
			datasetTables = append(datasetTables, tableInfo)
		}

		// Sort the tables by last modified time for this dataset.
		sort.Slice(datasetTables, func(i, j int) bool {
			return datasetTables[i].LastModifiedTime.After(datasetTables[j].LastModifiedTime)
		})
		datasets[dataset.DatasetID] = datasetTables
	}

	return datasets, nil
}
