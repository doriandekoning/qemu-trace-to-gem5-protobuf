#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>

static FILE *trace_fp;
#define HEADER_MAGIC_NUMBER 0xf2b177cb0aa429b4ULL
#define HEADER_VERSION 4

#define READ_LITERAL(Variable) {if(fread(&Variable, sizeof(Variable), 1, trace_fp) != 1) {return -1;}}

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

	READ_LITERAL(header)

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
		READ_LITERAL(type)
		//If type is 1 it is a trace record
		if(type==1) {
			break;
		}

		uint64_t eventid;
		READ_LITERAL(eventid)
		uint32_t length;
		READ_LITERAL(length)

		char *name = malloc(length + 1);
		name[length] = '\0';
		if(fread(name, sizeof(char) * length, 1, trace_fp) != 1) {
			printf("Error reading name\n");
			return -1;
		}
		free(name);

	}

	return 0;

}

int read_event() {
	//Read the header
	uint64_t event_id;
	READ_LITERAL(event_id)
	if(event_id != 75){
		printf("Event types other than 75 are not supported\n");
		return -1;
	}
	uint64_t timestamp;
	READ_LITERAL(timestamp);
	uint32_t rec_len;
	READ_LITERAL(rec_len);
	uint32_t  trace_pid;
	READ_LITERAL(trace_pid);

	//Read the 3 args (cpu, vaddr and info)
	uint64_t cpu;
	uint64_t vaddr;
	uint64_t info;
	READ_LITERAL(cpu)
	READ_LITERAL(vaddr)
	READ_LITERAL(info)

	printf("Read event timestamp: %llu pid:%lu cpu:%llx vaddr:%08X info:%lx\n",
	       timestamp, trace_pid, cpu, vaddr, info);
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

	for(int i= 0; i < 10; i++) {
		if(read_event() != 0) {
			return 1;
		}
		uint64_t entryType;
		READ_LITERAL(entryType)
		if(entryType != 1){
			printf("Found unexpectd mapping\n");
			return 1;
		}
	}
}
