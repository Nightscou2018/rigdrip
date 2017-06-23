package cmd

import (
	"log"

	"github.com/paypal/gatt"
	"github.com/spf13/cobra"
)

// collectorCmd represents the collector command
var collectorCmd = &cobra.Command{
	Use:   "collector",
	Short: "Process to collect G5 sensor readings",
	Long:  ``,
	Run:   Collector,
}

func init() {
	RootCmd.AddCommand(collectorCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// collectorCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// collectorCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

var TransmitterName = "DexcomTQ"

var DeviceInfo = gatt.MustParseUUID("0000180A-0000-1000-8000-00805F9B34FB")
var Advertisement = gatt.MustParseUUID("0000FEBC-0000-1000-8000-00805F9B34FB")
var CGMService = gatt.MustParseUUID("F8083532-849E-531C-C594-30F1F86A4EA5")
var ServiceB = gatt.MustParseUUID("F8084532-849E-531C-C594-30F1F86A4EA5")
var ManufacturerNameString = gatt.MustParseUUID("00002A29-0000-1000-8000-00805F9B34FB")
var CGMCommunication = gatt.MustParseUUID("F8083533-849E-531C-C594-30F1F86A4EA5")
var CGMControl = gatt.MustParseUUID("F8083534-849E-531C-C594-30F1F86A4EA5")
var CGMAuthentication = gatt.MustParseUUID("F8083535-849E-531C-C594-30F1F86A4EA5")
var CGMProbablyBackfill = gatt.MustParseUUID("F8083536-849E-531C-C594-30F1F86A4EA5")
var CharacteristicE = gatt.MustParseUUID("F8084533-849E-531C-C594-30F1F86A4EA5")
var CharacteristicF = gatt.MustParseUUID("F8084534-849E-531C-C594-30F1F86A4EA5")
var CharacteristicUpdateNotification = gatt.MustParseUUID("00002902-0000-1000-8000-00805F9B34FB")

var done = make(chan struct{})

func Collector(cmd *cobra.Command, args []string) {
	var DefaultClientOptions = []gatt.Option{
		gatt.LnxMaxConnections(1),
		gatt.LnxDeviceID(-1, false),
	}

	d, err := gatt.NewDevice(DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}

	d.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(onPeriphConnected),
		//gatt.PeripheralDisconnected(onPeriphDisconnected),
	)

	d.Init(onStateChanged)
	<-done
	log.Println("Done")
}

func onStateChanged(d gatt.Device, s gatt.State) {
	log.Println("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		log.Println("scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	log.Printf("Found device: %s", a.LocalName)
	if a.LocalName == TransmitterName {
		log.Printf("Dexcom G5 Discovered: %s \n", p.Name())
		p.Device().StopScanning()
		p.Device().Connect(p)
	}
}

var cgmServ *gatt.Service
var authChar *gatt.Characteristic
var commChar *gatt.Characteristic
var conChar *gatt.Characteristic

func onPeriphConnected(p gatt.Peripheral, err error) {
	log.Printf("Dextom G5 connected\n")

	services, err := p.DiscoverServices(nil)
	if err != nil {
		log.Printf("Failed to discover services, err: %s\n", err)
		return
	}

	for _, service := range services {
		if service.UUID().Equal(CGMService) {
			log.Printf("CGM Service Found %s\n", service.Name())
			cgmServ = service
			cs, _ := p.DiscoverCharacteristics(nil, service)
			for _, c := range cs {
				if c.UUID().Equal(CGMCommunication) {
					log.Print("Found CGM Communication")
					commChar = c
				} else if c.UUID().Equal(CGMControl) {
					log.Print("Found CGM Control")
					conChar = c
				} else if c.UUID().Equal(CGMAuthentication) {
					log.Print("Found CGM Auth")
					authChar = c

				} else if c.UUID().Equal(CGMProbablyBackfill) {
					log.Print("Found CGM Backfill")
				}
				/*if (c.UUID().Equal(uartServiceTXCharId)) {
				  log.Println("TX Characteristic Found")




				}*/
			}
		}
	}

	// Hopefully we have comm, con, and auth here...

	log.Print("Sending Auth Message")
	p.WriteCharacteristic(authChar, []byte{0x1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x2}, false)

	/*
		// read auth status...
		var res []byte
		res, _ = p.ReadCharacteristic(authChar)
		log.Print("Auth Read:")
		spew.Dump(res)
	*/

	p.DiscoverDescriptors(nil, conChar)
	p.SetNotifyValue(conChar, onControlNotify)

	/*

			 final BluetoothGattDescriptor descriptor = controlCharacteristic.getDescriptor(BluetoothServices.CharacteristicUpdateNotification);
		                descriptor.setValue(BluetoothGattDescriptor.ENABLE_INDICATION_VALUE);
		                if (useG5NewMethod()) {
		                    // new style
		                    GlucoseTxMessage glucoseTxMessage = new GlucoseTxMessage();
		                    controlCharacteristic.setValue(glucoseTxMessage.byteSequence);
		                } else {
		                    // old style
		                    SensorTxMessage sensorTx = new SensorTxMessage();
		                    controlCharacteristic.setValue(sensorTx.byteSequence);
		                }
		                Log.d(TAG,"getSensorData(): writing desccrptor");
		                mGatt.writeDescriptor(descriptor);


	*/

}

func onControlNotify(c *gatt.Characteristic, b []byte, e error) {
	log.Printf("CGM CONTROL Got back %s\n", string(b))
}
