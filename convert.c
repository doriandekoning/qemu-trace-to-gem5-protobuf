#include <stdio.h>
#include <stdint.h>

static FILE *trace_fp;


typedef struct {
    uint64_t header_event_id; /* HEADER_EVENT_ID */
    uint64_t header_magic;    /* HEADER_MAGIC    */
    uint64_t header_version;  /* HEADER_VERSION  */
} TraceLogHeader;

int read_header() {

	TraceLogHeader header = {
		.header_event_id  = 0,
		.header_magic  = 0,
		.header_version = 0,
	};


	if (fread(&header, sizeof header, 1, trace_fp) != 1) {
		return 1;
	}

	printf("Header event id: %llu\n", header.header_event_id);
	printf("Header magic: %llx\n", header.header_magic);
	printf("Header version: %llu\n", header.header_version);

	return 0;
}

int main(int argc, char *argv[]) {
	if(argc == 1 ){
		printf("First argument is the path of the trace input file to convert\n");
		return 1;
	}

	printf("Trace input file: %s\n", argv[1]);

	// Open file
	trace_fp = fopen(argv[1], "r");
	if (!trace_fp){
		printf("Could not open file\n");
		return -1;
	}

	if(read_header()) {
		printf("Unable to read header\n");
	}
	// Close file
	fclose(trace_fp);
}
