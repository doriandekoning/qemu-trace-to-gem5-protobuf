#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>

static FILE *trace_fp;
#define HEADER_MAGIC_NUMBER 0xf2b177cb0aa429b4ULL
#define HEADER_VERSION 4
typedef struct {
    uint64_t header_event_id; /* HEADER_EVENT_ID */
    uint64_t header_magic;    /* HEADER_MAGIC    */
    uint64_t header_version;  /* HEADER_VERSION  */
} TraceLogHeader;


typedef struct {
    uint64_t event; /* event ID value */
    uint64_t timestamp_ns;
    uint32_t length;   /*    in bytes */
    uint32_t pid;
    uint64_t arguments[];
} TraceRecord;


//Read header and print header
int read_header() {

	TraceLogHeader header = {
		.header_event_id  = 0,
		.header_magic  = 0,
		.header_version = 0,
	};


	if (fread(&header, sizeof(header), 1, trace_fp) != 1) {
		return -1;
	}

	printf("Header event id: %llu\n", header.header_event_id);
	printf("Header magic: 0x%llx\n", header.header_magic);
	printf("Header version: %llu\n", header.header_version);
	if(header.header_version != HEADER_VERSION ){
		printf("File has wrong header version!\n");
		return -1;
	}
	if(header.header_magic != HEADER_MAGIC_NUMBER ) {
		printf("Wrong header magic number\n");
		return -1;
	}
	return 0;
}

//Read mapping and print mapping
int read_mapping() {
	//Loop until a trace record is found
	while(1) {
		uint64_t type;
		if(fread(&type, sizeof(type), 1, trace_fp) != 1) {
			printf("Unable to read header\n");
			return -1;
		}
		//If type is 1 it is a trace record
		if(type==1) {
			break;
		}

		uint64_t eventid;
		if(fread(&eventid, sizeof(eventid), 1, trace_fp) != 1) {
			printf("Unable to read eventid\n");
			return -1;
		}
		uint32_t length;
		if(fread(&length, sizeof(length), 1, trace_fp) != 1) {
			printf("Unable to read length\n");
			return -1;
		}

		char *name = malloc(length + 1);
		name[length] = '\0';
		if(fread(name, sizeof(char) * length, 1, trace_fp) != 1) {
			printf("Error reading name\n");
			return -1;
		}
		printf("%llu:%s\n",eventid, name);
		free(name);

	}

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
		fclose(trace_fp);
		return 1;
	}
	if(read_mapping()) {
		printf("Unable to read mapping\n");
		fclose(trace_fp);
		return 1;
	}
}
