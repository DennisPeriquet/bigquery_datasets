.PHONY: all gen1 gen2 gen1a gen2a clean

LOG_FILE1 := openshift-ci-data-analysis.log
LOG_FILE2 := openshift-gce-devel.log
LOG_FILE1a := openshift-ci-data-analysis-mod.log
LOG_FILE2a := openshift-gce-devel-mod.log
EXECUTABLE := list_datasets

all: $(EXECUTABLE) gen1 gen2 gen1a gen2a

build: $(EXECUTABLE)

$(EXECUTABLE): cmd/list_datasets/main.go
	@go build -gcflags='-N -l' $(shell grep "module " go.mod | awk '{print $$2}')/cmd/list_datasets

$(LOG_FILE1): $(EXECUTABLE)
	./$(EXECUTABLE) openshift-ci-data-analysis ~/redhat_keys/openshift-ci-data-analysis-0ed2e02855f9-reader.json > $(LOG_FILE1)

$(LOG_FILE2): $(EXECUTABLE)
	./$(EXECUTABLE) openshift-gce-devel ~/redhat_keys/openshift-gce-devel-2245e68424fd.json > $(LOG_FILE2)

# The generated tables have last modified times and columns names.
#
gen1: $(LOG_FILE1)

gen2: $(LOG_FILE2)

# The tables ending in "-mod" don't have the columns in the report making it easy
# to scan the last modified times.
#
gen1a: gen1
	@echo "Printing out 'last modified' tables for openshift-ci-data-analysis..."
	grep -v "Description=" $(LOG_FILE1) > $(LOG_FILE1a)

gen2a: gen2
	@echo "Printing out 'last modified' tables for openshift-gce-devel..."
	grep -v "Description=" $(LOG_FILE2) > $(LOG_FILE2a)

clean:
	rm -f $(EXECUTABLE) $(LOG_FILE1) $(LOG_FILE2) $(LOG_FILE1a) $(LOG_FILE2a)