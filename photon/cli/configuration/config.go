package configuration

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

type TenantConfiguration struct {
	Name string
	ID   string
}

type ProjectConfiguration struct {
	Name string
	ID   string
}

type Configuration struct {
	CloudTarget       string
	Token             string
	IgnoreCertificate bool
	Tenant            *TenantConfiguration
	Project           *ProjectConfiguration
}

// Load configuration in config file
func LoadConfig() (*Configuration, error) {
	filepath, err := getConfigurationFilePath()
	if err != nil {
		return &Configuration{}, err
	}

	if isFileExist(filepath) {
		config, err := readConfigFromFile(filepath)
		if err != nil {
			return &Configuration{}, err
		}
		return config, nil
	}

	return &Configuration{}, nil
}

// Save configuration into config file, will overwrite config file
func SaveConfig(config *Configuration) error {
	filepath, err := getConfigurationFilePath()
	if err != nil {
		return err
	}

	err = writeConfigToFile(filepath, config)
	if err != nil {
		return err
	}

	return nil
}

func getUserConfigDirectory() (string, error) {
	var err error
	var homedir_input = "HOME"
	if runtime.GOOS == "windows" {
		homedir_input = "APPDATA"
	}
	userConfigDir := os.Getenv(homedir_input)
	userConfigDir = path.Join(userConfigDir, ".photon-cli")

	if isFileExist(userConfigDir) && !isFileDirectory(userConfigDir) {
		//there seems to be a file by this name
		//delete it - this is remnants from older versions
		err = os.Remove(userConfigDir)
		if err != nil {
			fmt.Println(err)
		}

	}
	//Ensure Config Dir Exists - if not create it
	if !isFileExist(userConfigDir) {
		err = os.Mkdir(userConfigDir, 0755)
		if err != nil {
			fmt.Println(err)
		}
	}

	return userConfigDir, err
}

// Get path of local config file: $HOME_DIR/.photon-cli
func getConfigurationFilePath() (string, error) {
	userConfigDir, err := getUserConfigDirectory()
	if err == nil {
		return path.Join(userConfigDir, ".photon-config"), nil
	}
	return userConfigDir, err
}

// Check if file is a directory
func isFileDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	} else {
		return fileInfo.IsDir()
	}
}

// Check if file exists
func isFileExist(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

// Read and deserialize configuration form local config file in JSON format
func readConfigFromFile(path string) (*Configuration, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error loading configuration: %v", err)
	}

	var config Configuration
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("Error loading configuration: %v", err)
	}

	return &config, nil
}

// Serialize and write configuration to local config file in JSON format
func writeConfigToFile(path string, config *Configuration) error {
	data, err := json.Marshal(*config)
	if err != nil {
		return fmt.Errorf("Error saving configuration: %v", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Error saving configuration: %v", err)
	}

	defer checkClose(&err, file)

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("Error saving configuration: %v", err)
	}

	return nil
}

func checkClose(errp *error, c io.Closer) {
	err := c.Close()
	if err != nil && *errp == nil {
		*errp = err
	}
}

func getCertsDir() (string, error) {
	var err error
	err = nil
	//Get Cert directory by adding .pc-certs to the HOME diretory of user
	certsDir, err := getUserConfigDirectory()
	if err != nil {
		return certsDir, err
	}
	certsDir = path.Join(certsDir, ".photon-cli-certs")

	//Ensure Certs Dir Exists - if not create it
	if !isFileExist(certsDir) {
		err := os.Mkdir(certsDir, 0755)
		if err != nil {
			fmt.Println(err)
		}
	}
	return certsDir, err
}

func generateFileNameFromCert(cert *x509.Certificate) (string, error) {
	//
	//Use the first 8 bytes of the Ceritificate's sha1 hash to
	//generate the file name for the cert
	sha1 := sha1.Sum(cert.Raw)
	certfilename := fmt.Sprintf("%x.pem", sha1[0:7])

	//Get the directory where all the certs are for this client
	certsDir, err := getCertsDir()
	if err == nil {
		certfilename = path.Join(certsDir, certfilename)
	}
	return certfilename, nil
}

func GetCertsFromLocalStore() (*x509.CertPool, error) {
	roots := x509.NewCertPool()

	//If we can't get the certs dir itself it's bad
	certsDir, err := getCertsDir()
	if err != nil {
		fmt.Println(err)
		return roots, err
	}
	files, err := ioutil.ReadDir(certsDir)

	//it's ok to not have any certs though
	for _, f := range files {
		//
		//If the file isn't a pem file ignore it
		if filepath.Ext(f.Name()) != ".pem" {
			continue
		}
		ca_b, err := ioutil.ReadFile(path.Join(certsDir, f.Name()))
		if err == nil {
			ca, err := x509.ParseCertificate(ca_b)
			if err == nil {
				roots.AddCert(ca)
			} else {
				fmt.Println(err)
			}
		} else {
			fmt.Println(err)
		}
	}
	return roots, err
}
func AddCertToLocalStore(cert *x509.Certificate) error {
	ca_f, err := generateFileNameFromCert(cert)

	if err == nil && !isFileExist(ca_f) {
		err = ioutil.WriteFile(ca_f, cert.Raw, 0644)
		if err != nil {
			fmt.Println(err)
		}
	}

	return err
}

func RemoveCertFromLocalStore(cert *x509.Certificate) error {
	ca_f, err := generateFileNameFromCert(cert)

	if err == nil && isFileExist(ca_f) {
		err = os.Remove(ca_f)
		if err != nil {
			fmt.Println(err)
		}
	}

	return err
}
