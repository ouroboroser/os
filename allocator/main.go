package main

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"
)

const (
	MEM_COMMIT  = 0x1000;
	MEM_RESERVE = 0x2000;
	MEM_RELEASE = 0x8000;
	PAGE_EXECUTE_READWRITE = 0x40;
	MEM = 10000;
)

var (
	kernel32     = syscall.MustLoadDLL("kernel32.dll")
	VirtualAlloc = kernel32.MustFindProc("VirtualAlloc");
	VirtualFree  = kernel32.MustFindProc("VirtualFree");
)

type Allocator struct {
	size uintptr;
	adress uintptr;
}

func (allocator Allocator) memAlloc() (uintptr, error) {
	addr, _, msg := VirtualAlloc.Call(0, allocator.size, MEM_RESERVE|MEM_COMMIT, PAGE_EXECUTE_READWRITE)
	if msg != nil {
		fmt.Println(msg)
		return addr, nil
	} else {
		return 0, nil
	}
}

func (allocator Allocator) requestMemory(memory*[MEM]byte, freeMemory int, headerSize int, blockSize int, startAddr * byte) (int, *byte) {
	size:=headerSize + blockSize;
	if freeMemory < size {
		fmt.Println("Not enough memory");
		return 0, nil
	} else {
		startIndex := 0;
		for i:=0; i<MEM; i ++ {
			if startAddr == &memory[i] {
				startIndex = i;
				break;
			}
		}
		freeMemory = freeMemory - size;
		fmt.Println("Update free memory", freeMemory);
		startAddr:=&memory[startIndex+size];
		return freeMemory, startAddr;
	}
	return freeMemory, startAddr;
}

func (allocator Allocator) freeAlloc() error {
	addr, _, msg := VirtualFree.Call(allocator.adress, 0 , MEM_RELEASE)
	if addr == 0 {
		fmt.Println(msg)
		return msg;
	} else {
		fmt.Println(msg)
		return nil
	}
}

type Header struct {
	size int;
	empty bool;
	current *byte;
	next *byte;
}

func (header *Header) Init(size int, empty bool, current *byte, next *byte ) {
	header.size = size;
	header.empty = empty;
	header.current = current;
	header.next = next;
}

func main()  {
	var newAllocator Allocator;
	newAllocator.size = MEM;
	addr, err:= newAllocator.memAlloc();
	if err != nil {
		log.Fatal(err)
	}
	newAllocator.adress = addr;
	fmt.Println(newAllocator);
	addrPointer := (*byte)(unsafe.Pointer(newAllocator.adress))
	if newAllocator.adress == 0 {
		fmt.Println("Adress: ", addrPointer);
	} else {
		memory := (*[MEM]byte)(unsafe.Pointer(addr));
		var freeMemory int;
		freeMemory = MEM;
		var addValue bool = true;
		for addValue {
			var blockSize int;
			fmt.Print("Enter value: ");
			fmt.Scan(&blockSize);
			header := new(Header);
			header.Init(blockSize, true, addrPointer, addrPointer)
			headerSize := unsafe.Sizeof(*header);
			convertHeaderSize:= (int(headerSize))
			updateFreeMemory, startAddr := newAllocator.requestMemory(memory, freeMemory, convertHeaderSize, blockSize, header.current);
			header.next = startAddr;
			addrPointer = startAddr;
			fmt.Printf("%+v\n", header);
			freeMemory = updateFreeMemory;
			if startAddr == nil {
				addValue = false;
			}
		}
		newAllocator.freeAlloc();
	}
}
