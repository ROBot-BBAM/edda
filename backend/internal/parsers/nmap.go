package parsers

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"

	"github.com/edda/backend/internal/database"
)

type NmapRun struct {
	XMLName xml.Name `xml:"nmaprun"`
	Hosts   []Host   `xml:"host"`
}

type Host struct {
	XMLName xml.Name `xml:"host"`
	Status  Status   `xml:"status"`
	Addresses []Address `xml:"address"`
	Hostnames []Hostname `xml:"hostnames>hostname"`
	OS       OS       `xml:"os"`
	Ports    Ports    `xml:"ports"`
}

type Status struct {
	State string `xml:"state,attr"`
}

type Address struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}

type Hostname struct {
	Name string `xml:"name,attr"`
}

type OS struct {
	OSMatch []OSMatch `xml:"osmatch"`
}

type OSMatch struct {
	Name     string `xml:"name,attr"`
	Accuracy string `xml:"accuracy,attr"`
}

type Ports struct {
	Port []Port `xml:"port"`
}

type Port struct {
	Protocol string  `xml:"protocol,attr"`
	PortID   int     `xml:"portid,attr"`
	State    State   `xml:"state"`
	Service  Service `xml:"service"`
}

type State struct {
	State string `xml:"state,attr"`
}

type Service struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr"`
	Version string `xml:"version,attr"`
}

func ParseNmapXML(filePath string, scanFileID uuid.UUID, db *database.DB) error {
	log.Printf("Parsing nmap XML file: %s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	var run NmapRun

	if err := decoder.Decode(&run); err != nil {
		return fmt.Errorf("failed to decode XML: %w", err)
	}

	log.Printf("Found %d hosts in nmap file", len(run.Hosts))

	for _, host := range run.Hosts {
		// Get IP address
		var ipAddress string
		for _, addr := range host.Addresses {
			if addr.AddrType == "ipv4" || addr.AddrType == "ipv6" {
				ipAddress = addr.Addr
				break
			}
		}

		if ipAddress == "" {
			log.Printf("Skipping host with no IP address")
			continue
		}

		// Get hostname
		var hostname string
		if len(host.Hostnames) > 0 {
			hostname = host.Hostnames[0].Name
		}

		// Get OS
		var os string
		if len(host.OS.OSMatch) > 0 {
			os = host.OS.OSMatch[0].Name
		}

		// Upsert host
		dbHost, err := db.UpsertHost(ipAddress, hostname, os, scanFileID)
		if err != nil {
			log.Printf("Failed to upsert host %s: %v", ipAddress, err)
			continue
		}

		// Process ports
		for _, port := range host.Ports.Port {
			// Only process open ports
			if port.State.State != "open" {
				continue
			}

			serviceName := port.Service.Name
			serviceProduct := port.Service.Product
			serviceVersion := port.Service.Version

			_, err := db.UpsertPort(
				dbHost.ID,
				port.PortID,
				port.Protocol,
				port.State.State,
				serviceName,
				serviceProduct,
				serviceVersion,
				scanFileID,
			)
			if err != nil {
				log.Printf("Failed to upsert port %d/%s on host %s: %v", port.PortID, port.Protocol, ipAddress, err)
				continue
			}
		}
	}

	log.Printf("Successfully parsed nmap XML file")
	return nil
}
