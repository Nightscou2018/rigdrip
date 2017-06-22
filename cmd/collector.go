// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/davecgh/go-spew/spew"
	"github.com/paypal/gatt"
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
	//spew.Dump(a)
	//spew.Dump(p)
	log.Printf("Found device: %s", a.LocalName)
	if a.LocalName == "DexcomTQ" {
		log.Printf("Dexcom G5 Discovered: %s \n", p.Name())
		p.Device().StopScanning()
		p.Device().Connect(p)
	}
}

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
			cs, _ := p.DiscoverCharacteristics(nil, service)
			for _, c := range cs {

				if c.UUID().Equal(CGMCommunication) {
					log.Print("Found CGM Communication")
				} else if c.UUID().Equal(CGMControl) {
					log.Print("Found CGM Control")
				} else if c.UUID().Equal(CGMAuthentication) {
					log.Print("Found CGM Auth")
				} else if c.UUID().Equal(CGMProbablyBackfill) {
					log.Print("Found CGM Backfill")
				} else {
					log.Print("Found UNKNOWN CHAR")
					spew.Dump(c.UUID())
				}

				//				spew.Dump(c.props())

				/*if (c.UUID().Equal(uartServiceTXCharId)) {
				  log.Println("TX Characteristic Found")

				  p.DiscoverDescriptors(nil, c)

				  p.SetNotifyValue(c, func(c *gatt.Characteristic, b []byte, e error) {
				      log.Printf("Got back %s\n", string(b))
				  })
				}*/
			}
		}
	}
}
