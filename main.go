package main

import (
	"math/rand"
	"sync"
	"time"
)

type CarConfig struct {
	Count          int
	ArrivalTimeMin int
	ArrivalTimeMax int
}
type StationConfig struct {
	Name         string
	Count        int
	ServeTimeMin int
	ServeTimeMax int
}
type RegistersConfig struct {
	Count         int
	HandleTimeMin int
	HandleTimeMax int
}
type Config struct {
	Cars     CarConfig
	Stations struct {
		Gas      StationConfig
		Diesel   StationConfig
		LPG      StationConfig
		Electric StationConfig
	}
	Registers RegistersConfig
}

type Car struct {
	ID            int
	ArrivalTime   int
	DepartureTime int
	QueueTime     int
	FuelType      StationConfig
	StationTime   int
	RegisterTime  int
}

var currentTime int

func randomDuration(min, max int) int {
	durationRange := time.Duration(max+1-min) * time.Millisecond
	randomDuration := time.Duration(rand.Int63n(int64(durationRange))) + time.Duration(min)*time.Millisecond

	return int(randomDuration.Milliseconds())
}

func randomFuelType(config Config) StationConfig {
	fuelInt := rand.Int63n(int64(4))
	switch fuelInt {
	case 0:
		return config.Stations.Gas
	case 1:
		return config.Stations.Diesel
	case 2:
		return config.Stations.LPG
	case 3:
		return config.Stations.Electric
	default:
		println("randomFuelType() error")
		return StationConfig{}
	}
}

func fillDuration(fuelType StationConfig) int {
	minServeTime := fuelType.ServeTimeMin
	maxServeTime := fuelType.ServeTimeMax
	return randomDuration(minServeTime, maxServeTime)
}

func gasWorker(gasQueue <-chan Car, registerQueue chan<- Car) {
	for car := range gasQueue {
		car.StationTime = fillDuration(car.FuelType)
		readTime := currentTime
		car.QueueTime = readTime - car.ArrivalTime
		registerQueue <- car
	}
}
func dieselWorker(dieselQueue <-chan Car, registerQueue chan<- Car) {
	for car := range dieselQueue {
		car.StationTime = fillDuration(car.FuelType)
		readTime := currentTime
		car.QueueTime = readTime - car.ArrivalTime
		registerQueue <- car
	}
}
func lpgWorker(lpgQueue <-chan Car, registerQueue chan<- Car) {
	for car := range lpgQueue {
		car.StationTime = fillDuration(car.FuelType)
		readTime := currentTime
		car.QueueTime = readTime - car.ArrivalTime
		registerQueue <- car
	}
}
func elecWorker(elecQueue <-chan Car, registerQueue chan<- Car) {
	for car := range elecQueue {
		car.StationTime = fillDuration(car.FuelType)
		readTime := currentTime
		car.QueueTime = readTime - car.ArrivalTime
		registerQueue <- car
	}
}
func registerWorker(config Config, registerQueue <-chan Car, results chan<- Car) {
	for car := range registerQueue {
		car.RegisterTime = randomDuration(config.Registers.HandleTimeMin, config.Registers.HandleTimeMax)
		readTime := currentTime
		car.DepartureTime = readTime + car.RegisterTime
		results <- car
	}
}

func main() {
	var wgServe sync.WaitGroup
	var wgRegister sync.WaitGroup
	config := Config{
		Cars: CarConfig{
			Count:          150000,
			ArrivalTimeMin: 1,
			ArrivalTimeMax: 2,
		},
		Stations: struct {
			Gas      StationConfig
			Diesel   StationConfig
			LPG      StationConfig
			Electric StationConfig
		}{
			Gas: StationConfig{
				Count:        2,
				ServeTimeMin: 2,
				ServeTimeMax: 5,
			},
			Diesel: StationConfig{
				Count:        2,
				ServeTimeMin: 3,
				ServeTimeMax: 6,
			},
			LPG: StationConfig{
				Count:        1,
				ServeTimeMin: 4,
				ServeTimeMax: 7,
			},
			Electric: StationConfig{
				Count:        1,
				ServeTimeMin: 5,
				ServeTimeMax: 10,
			},
		},
		Registers: RegistersConfig{
			Count:         2,
			HandleTimeMin: 1,
			HandleTimeMax: 3,
		},
	}
	// Create station queues
	gasQueue := make(chan Car, config.Cars.Count)
	dieselQueue := make(chan Car, config.Cars.Count)
	lpgQueue := make(chan Car, config.Cars.Count)
	elecQueue := make(chan Car, config.Cars.Count)
	registerQueue := make(chan Car, config.Cars.Count)
	results := make(chan Car, config.Cars.Count)

	currentTime := 0

	// Create go routines for all stations
	for i := 0; i < config.Stations.Gas.Count; i++ {
		go func() {
			wgServe.Add(1)
			gasWorker(gasQueue, registerQueue)
			wgServe.Done()
		}()
	}
	for i := 0; i < config.Stations.Diesel.Count; i++ {
		go func() {
			wgServe.Add(1)
			dieselWorker(dieselQueue, registerQueue)
			wgServe.Done()
		}()
	}
	for i := 0; i < config.Stations.LPG.Count; i++ {
		go func() {
			wgServe.Add(1)
			lpgWorker(lpgQueue, registerQueue)
			wgServe.Done()
		}()
	}
	for i := 0; i < config.Stations.Electric.Count; i++ {
		go func() {
			wgServe.Add(1)
			elecWorker(elecQueue, registerQueue)
			wgServe.Done()
		}()
	}
	for i := 0; i < config.Registers.Count; i++ {
		go func() {
			wgRegister.Add(1)
			registerWorker(config, registerQueue, results)
			wgRegister.Done()
		}()
	}

	// For every car in the simulation
	for i := 0; i < config.Cars.Count; i++ {
		newCarArrivalTime := randomDuration(config.Cars.ArrivalTimeMin, config.Cars.ArrivalTimeMax)
		currentTime += newCarArrivalTime

		car := Car{
			ID:          i + 1, // index from 1
			ArrivalTime: currentTime,
			FuelType:    randomFuelType(config),
		}

		// Assign each car to its fuel station queue
		switch car.FuelType {
		case config.Stations.Gas:
			gasQueue <- car
		case config.Stations.Diesel:
			dieselQueue <- car
		case config.Stations.LPG:
			lpgQueue <- car
		case config.Stations.Electric:
			elecQueue <- car
		default:
			println("Queue insert FuelType error")
		}

		// fuel queues get filled
		// fuel workers automatically move cars to register queue
	}
	close(gasQueue)    // workers can still pull from buffer when closed
	close(dieselQueue) // workers can still pull from buffer when closed
	close(lpgQueue)    // workers can still pull from buffer when closed
	close(elecQueue)   // workers can still pull from buffer when closed
	wgServe.Wait()
	println("Done serving fuel")
	close(registerQueue) // workers can still pull from buffer when closed
	wgRegister.Wait()
	println("Done the simulation")
	close(results) // workers can still pull from buffer when closed

	//Aggregate Results
	gasCars := 0
	gasTime := 0
	gasMaxQueueTime := 0
	dieselCars := 0
	dieselTime := 0
	dieselMaxQueueTime := 0
	lpgCars := 0
	lpgTime := 0
	lpgMaxQueueTime := 0
	elecCars := 0
	elecTime := 0
	elecMaxQueueTime := 0
	regCars := 0
	regTime := 0
	regMaxQueueTime := 0

	for j := range results {
		switch j.FuelType {
		case config.Stations.Gas:
			gasCars += 1
			gasTime += j.StationTime
			if j.StationTime > gasMaxQueueTime {
				gasMaxQueueTime = j.StationTime
			}
		case config.Stations.Diesel:
			dieselCars += 1
			dieselTime += j.StationTime
			if j.StationTime > dieselMaxQueueTime {
				dieselMaxQueueTime = j.StationTime
			}
		case config.Stations.LPG:
			lpgCars += 1
			lpgTime += j.StationTime
			if j.StationTime > lpgMaxQueueTime {
				lpgMaxQueueTime = j.StationTime
			}
		case config.Stations.Electric:
			elecCars += 1
			elecTime += j.StationTime
			if j.StationTime > elecMaxQueueTime {
				elecMaxQueueTime = j.StationTime
			}
		}
		regCars += 1
		regTime += j.RegisterTime
		if j.RegisterTime > regMaxQueueTime {
			regMaxQueueTime = j.RegisterTime
		}
	}
	gasAvgQueueTime := float64(gasTime) / float64(gasCars)
	dieselAvgQueueTime := float64(dieselTime) / float64(dieselCars)
	lpgAvgQueueTime := float64(lpgTime) / float64(lpgCars)
	elecAvgQueueTime := float64(elecTime) / float64(elecCars)
	regAvgQueueTime := float64(regTime) / float64(regCars)

	println("stations:")
	println("  gas:")
	println("    total_cars:", gasCars)
	println("    total_time:", gasTime/1000, "s")
	println("    avg_queue_time:", gasAvgQueueTime, "ms")
	println("    max_queue_time:", gasMaxQueueTime, "ms")
	println("  diesel:")
	println("    total_cars:", dieselCars)
	println("    total_time:", dieselTime/1000, "s")
	println("    avg_queue_time:", dieselAvgQueueTime, "ms")
	println("    max_queue_time:", dieselMaxQueueTime, "ms")
	println("  lpg:")
	println("    total_cars:", lpgCars)
	println("    total_time:", lpgTime/1000, "s")
	println("    avg_queue_time:", lpgAvgQueueTime, "ms")
	println("    max_queue_time:", lpgMaxQueueTime, "ms")
	println("  electric:")
	println("    total_cars:", elecCars)
	println("    total_time:", elecTime/1000, "s")
	println("    avg_queue_time:", elecAvgQueueTime, "ms")
	println("    max_queue_time:", elecMaxQueueTime, "ms")
	println("registers:")
	println("  total_cars:", regCars)
	println("  total_time:", regTime/1000, "s")
	println("  avg_queue_time:", regAvgQueueTime, "ms")
	println("  max_queue_time:", regMaxQueueTime, "ms")
}
