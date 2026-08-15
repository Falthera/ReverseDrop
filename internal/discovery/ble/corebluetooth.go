// ReverseDrop - Copyright (C) ReverseDrop Contributors
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

//go:build darwin

package ble

/*
#cgo LDFLAGS: -framework CoreBluetooth -framework Foundation -lobjc

#include <objc/message.h>
#include <objc/runtime.h>
#include <CoreBluetooth/CoreBluetooth.h>
#include <dispatch/dispatch.h>
#include <string.h>
#include <stdint.h>

typedef struct objc_class *Class;
typedef struct objc_object *id;
typedef struct objc_selector *SEL;
typedef void *dispatch_queue_t;

// Forward declare protocol so the compiler accepts @optional methods
@protocol CBCentralManagerDelegate <NSObject>
@optional
- (void)centralManagerDidUpdateState:(id)central;
- (void)centralManager:(id)central didDiscoverPeripheral:(id)peripheral
     advertisementData:(id)advertisementData RSSI:(id)RSSI;
- (void)centralManager:(id)central didConnectPeripheral:(id)peripheral;
- (void)centralManager:(id)central didFailToConnectPeripheral:(id)peripheral error:(id)error;
- (void)centralManager:(id)central didDisconnectPeripheral:(id)peripheral error:(id)error;
@end

// Go callback signature
typedef void (*GoScanCallback)(
    unsigned long long scannerID,
    const char *address,
    int rssi,
    const char *localName,
    const unsigned char *mfgData,
    unsigned int mfgDataLen,
    const char **serviceUUIDs,
    unsigned int serviceUUIDCount,
    const char **serviceDataKeys,
    const unsigned char **serviceDataValues,
    unsigned int *serviceDataLengths,
    unsigned int serviceDataCount,
    long long timestamp
);

// Global callback state
static GoScanCallback g_goScanCallback = NULL;
static unsigned long long g_scannerID = 0;

// Type-safe wrappers around objc_msgSend
static inline id msgSend_id(id self, SEL sel) {
    return ((id (*)(id, SEL))objc_msgSend)(self, sel);
}

static inline id msgSend_id_id(id self, SEL sel, id arg) {
    return ((id (*)(id, SEL, id))objc_msgSend)(self, sel, arg);
}

static inline id msgSend_id_id_id(id self, SEL sel, id a, id b) {
    return ((id (*)(id, SEL, id, id))objc_msgSend)(self, sel, a, b);
}

static inline void msgSend_void(id self, SEL sel) {
    ((void (*)(id, SEL))objc_msgSend)(self, sel);
}

static inline void msgSend_void_id(id self, SEL sel, id arg) {
    ((void (*)(id, SEL, id))objc_msgSend)(self, sel, arg);
}

static inline id objc_getClass(const char *name) {
    return ((id (*)(const char *))objc_getClass)(name);
}

static inline SEL sel_registerName(const char *name) {
    return ((SEL (*)(const char *))sel_registerName)(name);
}

static inline id class_createInstance(Class cls, size_t extra) {
    return class_createInstance(cls, extra);
}

static inline void class_addMethod(Class cls, SEL name, IMP imp, const char *types) {
    class_addMethod(cls, name, imp, types);
}

static inline void objc_registerClassPair(Class cls) {
    objc_registerClassPair(cls);
}

static inline id dispatch_queue_create(const char *label, id attr) {
    return dispatch_queue_create(label, attr);
}

// CBCentralManager lifecycle
static inline id CBCentralManager_init(id manager, id delegate, id queue) {
    SEL sel = sel_registerName("initWithDelegate:queue:");
    return ((id (*)(id, SEL, id, id))objc_msgSend)(manager, sel, delegate, queue);
}

// Scanning
static inline void CBCentralManager_scanForPeripherals(id mgr, id services, id options) {
    SEL sel = sel_registerName("scanForPeripheralsWithServices:options:");
    ((void (*)(id, SEL, id, id))objc_msgSend)(mgr, sel, services, options);
}

static inline void CBCentralManager_stopScan(id mgr) {
    SEL sel = sel_registerName("stopScan");
    ((void (*)(id, SEL))objc_msgSend)(mgr, sel);
}

// Connection
static inline void CBCentralManager_connect(id mgr, id peripheral, id options) {
    SEL sel = sel_registerName("connectPeripheral:options:");
    ((void (*)(id, SEL, id, id))objc_msgSend)(mgr, sel, peripheral, options);
}

static inline void CBCentralManager_cancelConnect(id mgr, id peripheral) {
    SEL sel = sel_registerName("cancelPeripheralConnection:");
    ((void (*)(id, SEL, id))objc_msgSend)(mgr, sel, peripheral);
}

// Peripheral properties
static inline id CBPeripheral_getName(id peripheral) {
    SEL sel = sel_registerName("name");
    return ((id (*)(id, SEL))objc_msgSend)(peripheral, sel);
}

static inline id CBPeripheral_getIdentifier(id peripheral) {
    SEL sel = sel_registerName("identifier");
    return ((id (*)(id, SEL))objc_msgSend)(peripheral, sel);
}

static inline id CBPeripheral_getState(id peripheral) {
    SEL sel = sel_registerName("state");
    return ((id (*)(id, SEL))objc_msgSend)(peripheral, sel);
}

// NSString helpers
static inline id NSString_fromUTF8(const char *cstr) {
    Class cls = objc_getClass("NSString");
    SEL sel = sel_registerName("stringWithUTF8String:");
    return ((id (*)(id, SEL, const char *))objc_msgSend)(cls, sel, cstr);
}

static inline const char *NSString_toUTF8(id nstring) {
    SEL sel = sel_registerName("UTF8String");
    return ((const char * (*)(id, SEL))objc_msgSend)(nstring, sel);
}

// NSUUID helpers
static inline id NSUUID_UUIDString(id uuid) {
    SEL sel = sel_registerName("UUIDString");
    return ((id (*)(id, SEL))objc_msgSend)(uuid, sel);
}

// NSNumber helpers
static inline int NSNumber_getInt(id number) {
    SEL sel = sel_registerName("intValue");
    return ((int (*)(id, SEL))objc_msgSend)(number, sel);
}

// NSDictionary helpers
static inline id NSDictionary_objectForKey(id dict, id key) {
    SEL sel = sel_registerName("objectForKey:");
    return ((id (*)(id, SEL, id))objc_msgSend)(dict, sel);
}

static inline unsigned long NSDictionary_count(id dict) {
    SEL sel = sel_registerName("count");
    return ((unsigned long (*)(id, SEL))objc_msgSend)(dict, sel);
}

static inline id NSDictionary_allKeys(id dict) {
    SEL sel = sel_registerName("allKeys");
    return ((id (*)(id, SEL))objc_msgSend)(dict, sel);
}

// NSArray helpers
static inline unsigned long NSArray_count(id arr) {
    SEL sel = sel_registerName("count");
    return ((unsigned long (*)(id, SEL))objc_msgSend)(arr, sel);
}

static inline id NSArray_objectAtIndex(id arr, unsigned long idx) {
    SEL sel = sel_registerName("objectAtIndex:");
    return ((id (*)(id, SEL, unsigned long))objc_msgSend)(arr, sel);
}

// NSData helpers
static inline unsigned long NSData_length(id data) {
    SEL sel = sel_registerName("length");
    return ((unsigned long (*)(id, SEL))objc_msgSend)(data, sel);
}

static inline const void *NSData_bytes(id data) {
    SEL sel = sel_registerName("bytes");
    return ((const void * (*)(id, SEL))objc_msgSend)(data, sel);
}

// CBAdvertisementData keys (NSString constants)
static inline id CBAdvertisementDataLocalNameKey(void) {
    return NSString_fromUTF8("kCBAdvDataLocalName");
}

static inline id CBAdvertisementDataManufacturerDataKey(void) {
    return NSString_fromUTF8("kCBAdvDataManufacturerData");
}

static inline id CBAdvertisementDataServiceUUIDsKey(void) {
    return NSString_fromUTF8("kCBAdvDataServiceUUIDs");
}

static inline id CBAdvertisementDataServiceDataKey(void) {
    return NSString_fromUTF8("kCBAdvDataServiceData");
}

// Apple company identifier for iBeacon / AirDrop detection
#define APPLE_COMPANY_ID 0x004C

// IMP cast helper
static inline IMP make_imp(void *fn) {
    return (IMP)fn;
}

// C helper to retrieve the exported Go callback function pointer.
// Using a C helper avoids cgo restrictions on referencing exported
// Go function pointers from Go code.
static inline GoScanCallback get_go_scan_callback(void) {
    return goScanCallbackEntry;
}

// C helper to set the active scanner ID for callbacks.
static inline void set_scanner_id(unsigned long long id) {
    g_scannerID = id;
}

// ============================================================================
// Objective-C Delegate Method Implementations
// These are called from the Objective-C runtime and forward to Go callbacks.
// ============================================================================

void centralManagerDidUpdateState(id self, SEL _cmd, id centralManager) {
    if (g_goScanCallback != NULL) {
        g_goScanCallback(g_scannerID, NULL, 0, NULL, NULL, 0, NULL, 0, NULL, NULL, NULL, 0, 0);
    }
}

void centralManagerDidDiscoverPeripheral(
    id self, SEL _cmd,
    id centralManager,
    id peripheral,
    id advertisementData,
    id rssiNumber
) {
    if (g_goScanCallback == NULL || peripheral == NULL) {
        return;
    }

    id identifier = CBPeripheral_getIdentifier(peripheral);
    const char *address = NULL;
    if (identifier != NULL) {
        id uuidStr = NSUUID_UUIDString(identifier);
        address = NSString_toUTF8(uuidStr);
    }
    if (address == NULL) {
        address = "";
    }

    int rssi = 0;
    if (rssiNumber != NULL) {
        rssi = NSNumber_getInt(rssiNumber);
    }

    const char *localNameC = NULL;
    unsigned int mfgLen = 0;
    const unsigned char *mfgBytes = NULL;
    unsigned int uuidCount = 0;
    const char **uuids = NULL;
    unsigned int svcDataCount = 0;
    const char **svcDataKeys = NULL;
    const unsigned char **svcDataValues = NULL;
    unsigned int *svcDataLens = NULL;
    long long timestamp = 0;

    if (advertisementData != NULL) {
        id localNameKey = CBAdvertisementDataLocalNameKey();
        id localNameObj = NSDictionary_objectForKey(advertisementData, localNameKey);
        if (localNameObj != NULL && localNameObj != objc_getClass("NSNull")) {
            localNameC = NSString_toUTF8(localNameObj);
        }
        if (localNameC == NULL) {
            localNameC = "";
        }

        id mfgKey = CBAdvertisementDataManufacturerDataKey();
        id mfgObj = NSDictionary_objectForKey(advertisementData, mfgKey);
        if (mfgObj != NULL && mfgObj != objc_getClass("NSNull")) {
            mfgLen = NSData_length(mfgObj);
            if (mfgLen > 0) {
                mfgBytes = NSData_bytes(mfgObj);
            }
        }

        id uuidKey = CBAdvertisementDataServiceUUIDsKey();
        id uuidArr = NSDictionary_objectForKey(advertisementData, uuidKey);
        if (uuidArr != NULL && uuidArr != objc_getClass("NSNull")) {
            uuidCount = NSArray_count(uuidArr);
            if (uuidCount > 0) {
                uuids = (const char **)malloc(uuidCount * sizeof(const char *));
                for (unsigned long i = 0; i < uuidCount; i++) {
                    id cbUUID = NSArray_objectAtIndex(uuidArr, i);
                    if (cbUUID != NULL && cbUUID != objc_getClass("NSNull")) {
                        id uuidStr = NSUUID_UUIDString(cbUUID);
                        uuids[i] = NSString_toUTF8(uuidStr);
                    } else {
                        uuids[i] = "";
                    }
                }
            }
        }

        id svcDataKey = CBAdvertisementDataServiceDataKey();
        id svcDataDict = NSDictionary_objectForKey(advertisementData, svcDataKey);
        if (svcDataDict != NULL && svcDataDict != objc_getClass("NSNull")) {
            id keys = NSDictionary_allKeys(svcDataDict);
            svcDataCount = NSArray_count(keys);
            if (svcDataCount > 0) {
                svcDataKeys = (const char **)malloc(svcDataCount * sizeof(const char *));
                svcDataValues = (const unsigned char **)malloc(svcDataCount * sizeof(const unsigned char *));
                svcDataLens = (unsigned int *)malloc(svcDataCount * sizeof(unsigned int));
                for (unsigned long i = 0; i < svcDataCount; i++) {
                    id cbUUID = NSArray_objectAtIndex(keys, i);
                    if (cbUUID != NULL && cbUUID != objc_getClass("NSNull")) {
                        id uuidStr = NSUUID_UUIDString(cbUUID);
                        svcDataKeys[i] = NSString_toUTF8(uuidStr);
                    } else {
                        svcDataKeys[i] = "";
                    }
                    id dataObj = NSDictionary_objectForKey(svcDataDict, cbUUID);
                    if (dataObj != NULL && dataObj != objc_getClass("NSNull")) {
                        svcDataLens[i] = NSData_length(dataObj);
                        if (svcDataLens[i] > 0) {
                            svcDataValues[i] = NSData_bytes(dataObj);
                        } else {
                            svcDataValues[i] = NULL;
                        }
                    } else {
                        svcDataLens[i] = 0;
                        svcDataValues[i] = NULL;
                    }
                }
            }
        }

        timestamp = (long long)time(NULL);
    }

    g_goScanCallback(
        g_scannerID,
        address,
        rssi,
        localNameC,
        mfgBytes,
        mfgLen,
        uuids,
        uuidCount,
        svcDataKeys,
        svcDataValues,
        svcDataLens,
        svcDataCount,
        timestamp
    );

    if (uuids != NULL) {
        free(uuids);
    }
    if (svcDataKeys != NULL) {
        free(svcDataKeys);
    }
    if (svcDataValues != NULL) {
        free(svcDataValues);
    }
    if (svcDataLens != NULL) {
        free(svcDataLens);
    }
}

void centralManagerDidConnectPeripheral(id self, SEL _cmd, id centralManager, id peripheral) {
    if (g_goScanCallback != NULL) {
        g_goScanCallback(g_scannerID, NULL, 0, NULL, NULL, 0, NULL, 0, NULL, NULL, NULL, 0, 0);
    }
}

void centralManagerDidFailToConnect(id self, SEL _cmd, id centralManager, id peripheral, id error) {
    if (g_goScanCallback != NULL) {
        g_goScanCallback(g_scannerID, NULL, 0, NULL, NULL, 0, NULL, 0, NULL, NULL, NULL, 0, 0);
    }
}

void centralManagerDidDisconnectPeripheral(id self, SEL _cmd, id centralManager, id peripheral, id error) {
    if (g_goScanCallback != NULL) {
        g_goScanCallback(g_scannerID, NULL, 0, NULL, NULL, 0, NULL, 0, NULL, NULL, NULL, 0, 0);
    }
}
*/
import "C"

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// appleCompanyID is the Apple company identifier (0x004C) used in BLE manufacturer data.
const appleCompanyID uint16 = 0x004C

// scannerIDCounter generates unique IDs for scanner instances.
// These IDs are passed to C callbacks to avoid storing Go pointers in C memory.
var scannerIDCounter uint64

// scannerMap maps scanner IDs to scanner instances.
var scannerMap sync.Map

// coreBluetoothScanner implements a real CBCentralManager-based BLE scanner.
//
// BUILD NOTES:
//   - This file is tagged with //go:build darwin and only compiles on macOS.
//   - Requires Xcode command-line tools for CoreBluetooth and Foundation headers.
//   - Links against CoreBluetooth, Foundation, and libobjc frameworks.
//
// MACOS TEST CYCLE REQUIRED (TODO: verify on macOS before production use):
//   1. Build and run on darwin/amd64 (Intel) and darwin/arm64 (Apple Silicon).
//   2. Verify objc_msgSend calling conventions on arm64:
//      - arm64 requires exact function signatures for msgSend variants;
//        variadic casts used here may trap on arm64 if the compiler
//        doesn't generate the correct veneer. If build fails on arm64,
//        add explicit non-variadic wrapper functions in the cgo section.
//   3. Verify CBCentralManagerDelegate protocol conformance check succeeds;
//      if the @optional protocol methods are not dispatched, ensure the
//      delegate class is created with objc_allocateClassPair and the
//      protocol is added via class_addProtocol before objc_registerClassPair.
//   4. Verify delegate callback bridge: GoScanCallback must be invoked
//      from every delegate method. The callback signature uses C types
//      so the Go runtime is not entered from Objective-C signal handlers.
//   5. Verify centralManagerDidUpdateState: fires with
//      CBCentralManagerStatePoweredOn before scanForPeripheralsWithServices
//      is called. Scanning before powered on is undefined behavior.
//   6. Verify dispatch queue lifecycle: the queue must remain valid for the
//      entire lifetime of the CBCentralManager. Do not release it early.
//   7. Verify Bluetooth permission flow on macOS 13+:
//      CBCentralManager will show a system dialog on first scan. If the
//      user denies, centralManagerDidUpdateState: returns
//      CBCentralManagerStateUnauthorized. Handle gracefully.
//   8. Verify advertisement data parsing for all keys:
//      kCBAdvDataLocalName, kCBAdvDataManufacturerData,
//      kCBAdvDataServiceUUIDs, kCBAdvDataServiceData.
//   9. Verify manufacturer data parsing: Apple company ID (0x004C) must be
//      detected at bytes 0-1 (little-endian) for AirDrop/iBeacon frames.
//   10. Verify memory management: cgo assumes ARC is enabled. If building
//       with MRR, add CFRetain/CFRelease for Objective-C objects crossing
//       the cgo boundary.
//   11. Verify stopScan is idempotent and safe to call multiple times.
//   12. Verify context cancellation triggers stopScan synchronously.
type coreBluetoothScanner struct {
	mu        sync.Mutex
	running   bool
	ch        chan Advertisement
	cancel    context.CancelFunc
	manager   unsafe.Pointer // *CBCentralManager (Objective-C id)
	delegate  unsafe.Pointer // delegate object implementing CBCentralManagerDelegate
	queue     unsafe.Pointer // dispatch_queue_t
	id        uint64         // unique ID for callback bridge
}

// NewCoreBluetoothScanner creates a new CoreBluetooth BLE scanner.
func NewCoreBluetoothScanner() Scanner {
	return &coreBluetoothScanner{ch: make(chan Advertisement, 64)}
}

// Scan starts BLE scanning and returns a channel that receives advertisements.
func (c *coreBluetoothScanner) Scan(ctx context.Context) (<-chan Advertisement, error) {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return c.ch, nil
	}
	c.mu.Unlock()

	if err := c.initializeManager(); err != nil {
		return nil, fmt.Errorf("corebluetooth: failed to initialize CBCentralManager: %w", err)
	}

	scanCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	c.mu.Lock()
	c.running = true
	c.mu.Unlock()

	go c.scanLoop(scanCtx)

	return c.ch, nil
}

// Stop stops BLE scanning and releases CoreBluetooth resources.
func (c *coreBluetoothScanner) Stop() error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	if c.cancel != nil {
		c.cancel()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.manager != nil {
		C.CBCentralManager_stopScan(c.manager)
	}

	scannerMap.Delete(c.id)
	c.running = false
	return nil
}

func (c *coreBluetoothScanner) scanLoop(ctx context.Context) {
	defer close(c.ch)

	<-ctx.Done()

	c.mu.Lock()
	c.running = false
	c.mu.Unlock()
}

// initializeManager creates the CBCentralManager and delegate if not already created.
// Must be called with c.mu NOT held (it acquires the lock internally).
func (c *coreBluetoothScanner) initializeManager() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.manager != nil {
		return nil
	}

	// Assign a unique ID for the callback bridge.
	// We use an atomic counter to generate IDs without race conditions.
	c.id = atomic.AddUint64(&scannerIDCounter, 1)
	scannerMap.Store(c.id, (*coreBluetoothScanner)(unsafe.Pointer(c)))

	// Register Go callback bridge so Objective-C delegate methods can call back
	// into Go. We use a C helper to retrieve the function pointer of the
	// //export'd Go function, avoiding cgo restrictions on direct references.
	C.g_goScanCallback = C.get_go_scan_callback()

	// Set the scanner ID so C callbacks know which Go scanner to dispatch to.
	C.set_scanner_id(C.ulonglong(c.id))

	// Create dispatch queue for delegate callbacks
	// TODO(macos-test): Verify queue label format accepted by Apple's dispatch APIs
	queueLabel := C.CString("com.reversedrop.ble.scanner")
	if queueLabel == nil {
		return fmt.Errorf("corebluetooth: failed to allocate dispatch queue label")
	}
	defer C.free(unsafe.Pointer(queueLabel))

	c.queue = unsafe.Pointer(C.dispatch_queue_create(queueLabel, C.NULL))
	if c.queue == nil {
		return fmt.Errorf("corebluetooth: failed to create dispatch queue")
	}

	// Create delegate class at runtime
	delegateClassName := C.CString("ReverseDropBLEDelegate")
	if delegateClassName == nil {
		return fmt.Errorf("corebluetooth: failed to allocate delegate class name")
	}
	defer C.free(unsafe.Pointer(delegateClassName))

	delegateClass := C.objc_allocateClassPair(
		C.objc_getClass("NSObject"),
		delegateClassName,
		0,
	)
	if delegateClass == 0 {
		return fmt.Errorf("corebluetooth: failed to allocate delegate class")
	}

	// Register delegate methods using class_addMethod
	// Each method is implemented in the cgo section above.
	// The type encoding string describes the method signature:
	//   v = void return
	//   @ = id (self)
	//   : = SEL (_cmd)
	//   @ = id (argument)
	C.class_addMethod(
		delegateClass,
		C.sel_registerName(C.CString("centralManagerDidUpdateState:")),
		C.make_imp(C.centralManagerDidUpdateState),
		C.CString("v@:@"),
	)
	C.class_addMethod(
		delegateClass,
		C.sel_registerName(C.CString("centralManager:didDiscoverPeripheral:advertisementData:RSSI:")),
		C.make_imp(C.centralManagerDidDiscoverPeripheral),
		C.CString("v@:@@@@"),
	)
	C.class_addMethod(
		delegateClass,
		C.sel_registerName(C.CString("centralManager:didConnectPeripheral:")),
		C.make_imp(C.centralManagerDidConnectPeripheral),
		C.CString("v@:@@"),
	)
	C.class_addMethod(
		delegateClass,
		C.sel_registerName(C.CString("centralManager:didFailToConnectPeripheral:error:")),
		C.make_imp(C.centralManagerDidFailToConnect),
		C.CString("v@:@@@"),
	)
	C.class_addMethod(
		delegateClass,
		C.sel_registerName(C.CString("centralManager:didDisconnectPeripheral:error:")),
		C.make_imp(C.centralManagerDidDisconnectPeripheral),
		C.CString("v@:@@@"),
	)

	C.objc_registerClassPair(delegateClass)

	// Allocate delegate instance
	c.delegate = unsafe.Pointer(C.class_createInstance(delegateClass, 0))
	if c.delegate == nil {
		return fmt.Errorf("corebluetooth: failed to allocate delegate instance")
	}

	// Allocate CBCentralManager instance
	cls := C.objc_getClass("CBCentralManager")
	if cls == nil {
		return fmt.Errorf("corebluetooth: CBCentralManager class not found")
	}

	allocSel := C.sel_registerName("alloc")
	c.manager = unsafe.Pointer(C.msgSend_id(cls, allocSel))
	if c.manager == nil {
		return fmt.Errorf("corebluetooth: failed to allocate CBCentralManager")
	}

	// initWithDelegate:queue:
	initSel := C.sel_registerName("initWithDelegate:queue:")
	c.manager = unsafe.Pointer(C.CBCentralManager_init(
		(*C.id)(c.manager),
		(*C.id)(c.delegate),
		(*C.id)(c.queue),
	))
	if c.manager == nil {
		return fmt.Errorf("corebluetooth: CBCentralManager initWithDelegate:queue: returned nil")
	}

	// Start scanning for all services (nil) with no options (nil)
	C.CBCentralManager_scanForPeripherals((*C.id)(c.manager), C.NULL, C.NULL)
	// TODO(macos-test): Verify scanForPeripheralsWithServices:nil options:nil is valid;
	//                    on some macOS versions, passing nil for services scans for all.

	return nil
}

// processAdvertisement parses CoreBluetooth advertisement data and emits an Advertisement.
func (c *coreBluetoothScanner) processAdvertisement(address, localName string, rssi int, mfgData map[uint16][]byte, serviceUUIDs []string, serviceData map[string][]byte) {
	select {
	case c.ch <- Advertisement{
		Address:          address,
		RSSI:             rssi,
		LocalName:        localName,
		ManufacturerData: mfgData,
		ServiceUUIDs:     serviceUUIDs,
		ServiceData:      serviceData,
		Timestamp:        time.Now().Unix(),
	}:
	default:
	}
}

// parseAdvertisementData extracts BLE advertisement fields from an NSDictionary.
//
// This is the pure-Go parsing logic. It is called from the Objective-C delegate
// callback bridge (see centralManagerDidDiscoverPeripheral in the cgo section).
//
// MACOS TEST CYCLE NOTES:
//   - NSDictionary keys for advertisement data are NSString objects.
//   - kCBAdvDataManufacturerData value is NSData containing raw bytes.
//   - kCBAdvDataServiceUUIDs value is NSArray of CBUUID objects.
//   - kCBAdvDataServiceData value is NSDictionary<CBUUID, NSData>.
//   - kCBAdvDataLocalName value is NSString.
func parseAdvertisementData(rawDict unsafe.Pointer) (localName string, mfgData map[uint16][]byte, serviceUUIDs []string, serviceData map[string][]byte) {
	if rawDict == nil {
		return "", nil, nil, nil
	}

	dict := (*C.id)(rawDict)
	localName = ""
	mfgData = make(map[uint16][]byte)
	serviceUUIDs = nil
	serviceData = make(map[string][]byte)

	// --- Local Name ---
	localNameKey := C.CBAdvertisementDataLocalNameKey()
	nameObj := C.NSDictionary_objectForKey(*dict, localNameKey)
	if nameObj != nil && nameObj != C.objc_getClass("NSNull") {
		localName = C.GoString(C.NSString_toUTF8(nameObj))
	}

	// --- Manufacturer Data ---
	mfgKey := C.CBAdvertisementDataManufacturerDataKey()
	mfgObj := C.NSDictionary_objectForKey(*dict, mfgKey)
	if mfgObj != nil && mfgObj != C.objc_getClass("NSNull") {
		dataLen := int(C.NSData_length(mfgObj))
		if dataLen >= 2 {
			bytes := C.NSData_bytes(mfgObj)
			raw := C.GoBytes(unsafe.Pointer(bytes), C.int(dataLen))
			// Manufacturer ID is little-endian at bytes 0-1
			companyID := uint16(raw[0]) | uint16(raw[1])<<8
			// TODO(macos-test): Verify byte order is little-endian on all macOS versions
			mfgData[companyID] = raw
		}
	}

	// --- Service UUIDs ---
	uuidKey := C.CBAdvertisementDataServiceUUIDsKey()
	uuidArr := C.NSDictionary_objectForKey(*dict, uuidKey)
	if uuidArr != nil && uuidArr != C.objc_getClass("NSNull") {
		count := int(C.NSArray_count(uuidArr))
		serviceUUIDs = make([]string, 0, count)
		for i := 0; i < count; i++ {
			cbUUID := C.NSArray_objectAtIndex(uuidArr, C.ulong(i))
			if cbUUID != nil && cbUUID != C.objc_getClass("NSNull") {
				uuidStrObj := C.NSUUID_UUIDString(cbUUID)
				uuidStr := C.GoString(C.NSString_toUTF8(uuidStrObj))
				serviceUUIDs = append(serviceUUIDs, uuidStr)
			}
		}
	}

	// --- Service Data ---
	svcDataKey := C.CBAdvertisementDataServiceDataKey()
	svcDataDict := C.NSDictionary_objectForKey(*dict, svcDataKey)
	if svcDataDict != nil && svcDataDict != C.objc_getClass("NSNull") {
		innerCount := int(C.NSDictionary_count(svcDataDict))
		if innerCount > 0 {
			keys := C.NSDictionary_allKeys(svcDataDict)
			keyCount := int(C.NSArray_count(keys))
			serviceData = make(map[string][]byte, keyCount)
			for i := 0; i < keyCount; i++ {
				cbUUID := C.NSArray_objectAtIndex(keys, C.ulong(i))
				if cbUUID != nil && cbUUID != C.objc_getClass("NSNull") {
					uuidStrObj := C.NSUUID_UUIDString(cbUUID)
					uuidStr := C.GoString(C.NSString_toUTF8(uuidStrObj))
					dataObj := C.NSDictionary_objectForKey(svcDataDict, cbUUID)
					if dataObj != nil && dataObj != C.objc_getClass("NSNull") {
						dataLen := int(C.NSData_length(dataObj))
						if dataLen > 0 {
							dataBytes := C.NSData_bytes(dataObj)
							serviceData[uuidStr] = C.GoBytes(unsafe.Pointer(dataBytes), C.int(dataLen))
						}
					}
				}
			}
		}
	}

	return localName, mfgData, serviceUUIDs, serviceData
}

//export goScanCallbackEntry
func goScanCallbackEntry(
	scannerID C.ulonglong,
	address *C.char,
	rssi C.int,
	localName *C.char,
	mfgData *C.uchar,
	mfgDataLen C.uint,
	serviceUUIDs **C.char,
	serviceUUIDCount C.uint,
	serviceDataKeys **C.char,
	serviceDataValues **C.uchar,
	serviceDataLengths *C.uint,
	serviceDataCount C.uint,
	timestamp C.longlong,
) {
	val, ok := scannerMap.Load(uint64(scannerID))
	if !ok {
		return
	}
	scanner, ok := val.(*coreBluetoothScanner)
	if !ok || scanner == nil {
		return
	}

	// Skip empty state updates (centralManagerDidUpdateState with no data)
	if address == nil && rssi == 0 && localName == nil && mfgData == nil &&
		serviceUUIDs == nil && serviceDataKeys == nil && serviceDataValues == nil {
		return
	}

	var addr string
	if address != nil {
		addr = C.GoString(address)
	}

	var name string
	if localName != nil {
		name = C.GoString(localName)
	}

	mfg := make(map[uint16][]byte)
	if mfgData != nil && mfgDataLen > 0 {
		raw := C.GoBytes(unsafe.Pointer(mfgData), C.int(mfgDataLen))
		companyID := uint16(raw[0]) | uint16(raw[1])<<8
		mfg[companyID] = raw
	}

	uuids := make([]string, 0, int(serviceUUIDCount))
	if serviceUUIDs != nil && serviceUUIDCount > 0 {
		ptrSize := unsafe.Sizeof(uintptr(0))
		for i := 0; i < int(serviceUUIDCount); i++ {
			ptr := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(serviceUUIDs)) + uintptr(i)*ptrSize))
			if ptr != nil {
				uuids = append(uuids, C.GoString(ptr))
			}
		}
	}

	svcData := make(map[string][]byte)
	if serviceDataKeys != nil && serviceDataValues != nil && serviceDataLengths != nil && serviceDataCount > 0 {
		ptrSize := unsafe.Sizeof(uintptr(0))
		lenSize := unsafe.Sizeof(C.uint(0))
		for i := 0; i < int(serviceDataCount); i++ {
			keyPtr := *(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(serviceDataKeys)) + uintptr(i)*ptrSize))
			if keyPtr == nil {
				continue
			}
			key := C.GoString(keyPtr)
			valPtr := *(*unsafe.Pointer)(unsafe.Pointer(uintptr(unsafe.Pointer(serviceDataValues)) + uintptr(i)*ptrSize))
			length := *(*C.uint)(unsafe.Pointer(uintptr(unsafe.Pointer(serviceDataLengths)) + uintptr(i)*lenSize))
			if valPtr != nil && length > 0 {
				svcData[key] = C.GoBytes(valPtr, C.int(length))
			}
		}
	}

	scanner.processAdvertisement(addr, name, int(rssi), mfg, uuids, svcData)
}
