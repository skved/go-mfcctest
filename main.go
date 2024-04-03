package main

import (
	"fmt"
	"math"
	"os"

	// "strings"

	// Temporary for testing output to file through string builder. Remove when done.
	"github.com/go-audio/wav"
	//"golang.org/x/exp/slices" // Import? or just manually do it?
	"gonum.org/v1/gonum/dsp/fourier" // "gonum"
	"gonum.org/v1/gonum/dsp/window"  // "gonum
)

// Load audio file
// Apply Pre-emphasis
// Amplify the high frequencies
// Frame the Signal
// Create overlapping segments (overlapping reduces artifacting?)
// Apply a Window Function - Like hamming
// Each frame get "windowed" by multiplying it with a window function
// Apply FFT to each windowed frame
// Converts from time domain to frequency domain
// Apply Triangular Filtering - A type of mel filter, so still spaced on the mel scale.
// The filter banks (triangular filtering) divide the signal's frequency spectrum into multiple frequency bands.
// So that each band can be analyzed seperately.
// Apply a logarithm to the filter bank energies
//Filter bank energies - computed by summing the magnitudes of the FFt output within each of the triangular filters
// Take the DCT - Discrete cosine transform
// Take the DCT of the list of mel log powers
// The MFCCs are the amplitudes of the resulting spectrum

const preEmphasis = 0.95 // Modify coefficient as needed

func main() {
	// Variables needed?
	audiofile, err := os.Open("test.wav") // "os"
	if err != nil {
		// need to put this in a function so that I can have a proper return type
		// os.open is forcing an error return type
		print("Error opening file")
	}

	decoder := wav.NewDecoder(audiofile) // "wav" // decoder // This should automatically
	if decoder == nil {
		print("Error decoding file")
	}
	audiobuffer, err := decoder.FullPCMBuffer() // "wav"
	if err != nil {
		print("why is everything broken")
	}
	// FullPCMBuffer takes a pointer to a decoder
	// returns a buffer
	// PCMBuffer is more efficient

	// fmt.Println(audiobuffer.Format.NumChannels)
	// fmt.Println(audiobuffer.Format.SampleRate) // Sample Rate will be found in aeneas? Or we find an "ideal" value independently? Or we need to match rates etc, to the provided audio?
	//fmt.Println(audiobuffer.Data)              // sin wave data
	audiofile.Close()

	// fmt.Println(audiobuffer.Data) // sin wave data

	// End Signal Loading
	// Start Pre Emphasis

	signal := audiobuffer.Data

	// Convert to float
	// SEE IF THERE IS A BETTER WAY
	// THERE ARE 500000 values!!
	signal64 := make([]float64, len(signal))
	for i := 0; i < len(signal); i++ {
		signal64[i] = float64(signal[i])
	}
	// End
	for i := 1; i < len(signal); i++ {
		signal64[i] = signal64[i] - preEmphasis*signal64[i-1]
	}

	// fmt.Println(len(signal))

	// End Pre-Emphasis
	// Do we need to normalize the signal? Wikipedia seems to indicate that it is necessary for windowing.
	// Start Normalization
	// Find the biggest value in the signal and divide all other values by it.

	currentMax := signal64[0]
	if currentMax < 0 {
		currentMax = currentMax * -1
	}

	// I think this will accurately find the largest value but have someone double check for errors.
	for i := 1; i < len(signal64); i++ {
		next := signal64[i]
		if next < 0 {
			next = next * -1
		}
		if next > currentMax {
			currentMax = next
		}
	}
	absMax := currentMax

	for i := 0; i < len(signal); i++ {
		signal64[i] = signal64[i] / absMax
	}

	// fmt.Println(signal64)

	// End Normalization
	// Start Framing the Signal

	// How much overlap is desired
	// Slice of slices
	// Orginial signal slice
	// Each new slice to put in the slice of slices

	// 44100(samplingRate) * 0.03(milliseconds) = 1323 (frame size)

	// Sampling Rate = 44.1kHz
	// Frame Length = 20-30ms probably is a good choice

	var sampleRate = 44100.0
	var frameLength = 0.03
	var frameOverlap = 0.5
	var frameSize = int(sampleRate * frameLength)                                        // int: can't have a fraction of a sample
	var frameStep = int(sampleRate * frameLength * frameOverlap)                         // int: indexes are whole numbers
	var numFrames = (float64(len(signal64)) / (sampleRate * frameLength)) / frameOverlap // +1 if partial
	if numFrames != float64(int(numFrames)) {                                            // If there is a partial frame truncate and add 1, then handle the the partial/tail frame.
		numFrames = float64(int(numFrames))
	}
	fmt.Println("signal64: ", len(signal64))
	fmt.Println("frameSize: ", frameSize)
	fmt.Println("frameStep: ", frameStep)
	fmt.Println("numFrames: ", numFrames)

	// So an overlap variable might be to multiply by the frame length by desired overlap. So for a 50% overlap, multiply framelength by 0.5. So, for [i] would be for [i+0.5*frameLength]

	//for (i := 0; i < numFrames; i++) {

	// frames := slices.Slice(signal64, 1323, 1323) // "slices"

	frames := make([][]float64, int(numFrames))
	for i := 0; i < int(numFrames); i++ {
		start := i * frameStep
		end := start + frameSize
		if end > len(signal64) {
			end = len(signal64)
		}
		frames[i] = signal64[start:end]
	}
	///////////////////////////////////////////////////////////////////////
	// Testing Output to File to use in online array visualizer
	// Convert the slice of slices to a string representation
	// var sb strings.Builder
	// for _, row := range frames {
	// 	//fmt.Fprintf(&sb, "%v\n", row)
	// 	fmt.Fprintf(&sb, "%s\n", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(row)), ","), "[]"))
	// }
	// data := sb.String()

	// os.WriteFile("frames.txt", []byte(fmt.Sprint(data)), 0644)
	// fmt.Println("Finished Writing to File")
	///////////////////////////////////////////////////////////////////////
	// Count the samples per frame then add them all up and make sure they equal the starting length of the signal.
	// var count int
	// for i := 0; i < len(frames); i++ {
	// 	// fmt.Println(len(frames[i]))
	// 	count += len(frames[i])
	// 	if i == len(frames)-2 {
	// 		fmt.Println("Count: ", count)
	// 		fmt.Println("2ndtolast: ", len(frames[i]))
	// 	}
	// 	if i == len(frames)-1 {
	// 		fmt.Println("Count: ", count)
	// 		fmt.Println("lastframe: ", len(frames[i]))
	// 	}
	// 	// Print the first overlap index of each frame
	// 	if i != 0 {
	// 		fmt.Println("Previous Frame: ", frames[i-1][660:663])
	// 		fmt.Println("Current Frame : ", frames[i][0:3])
	// 	}
	///////////////////////////////////////////////////////////////////////
	// Check to see if the last frames align. Hopefully this means that I've gathered all the data correctly.
	// fmt.Println(len(frames))
	// fmt.Println(frames[796][1257])
	// fmt.Println(signal64[len(signal64)-1])
	///////////////////////////////////////////////////////////////////////

	// End Framing the Signal
	// Start Windowing the Signal (Hamming Window)
	// https://pkg.go.dev/gonum.org/v1/gonum/dsp/window#example-Hamming
	// I believe the windowing function is applied inplace, so I need to create the slice of slices beforehand otherwise the windowing would change the values, which is a problem since there is going to be overlap and the values will need to be resued for multiple frames and we don't want it to get "windowed" multiple times?

	// Have to copy the frames because I need both the original and windowed frames for the FFT.

	goingHam := make([][]float64, len(frames))
	for i := 0; i < len(frames); i++ {
		goingHam[i] = append(goingHam[i], frames[i]...)
	}

	// Maybe run a quick test to see if they are equivalent, then remove it.
	// eg, for i in range, if frames[i] == goingHam[i] then print "same" else print "different"
	// HAVE NOT YET RUN THIS TEST

	for i := 0; i < len(goingHam); i++ {
		window.Hamming(goingHam[i]) // Changes data in place according to documentation
	}
	// Now the overlap-add step. https://en.wikipedia.org/wiki/Overlap%E2%80%93add_method
	// recombining the frames accounting for the overlap.

	// - FFT then overlap-add
	// - Overlap-add method
	// There doesn't seem to be a library for this, so, like framing, do it manually.
	// -

	// So even though the slice of slices are independent from each other and so there isn't an issue with the windowing function changing the values of the other frames, I still need to account for the overlap?

	// End Windowing the Signal

	// The DCT (or mdct) happens last

	// The FFT returns complex numbers, but Do I need a complex input?

	fftCoefficients := make([][]complex128, len(frames))

	for i := 0; i < len(frames); i++ {
		fft := fourier.NewFFT(len(frames[i]))
		fftCoefficients[i] = fft.Coefficients(nil, goingHam[i])
		//fmt.Println("Coefficients: ", fftCoefficients[i])
		//fmt.Println("FFT Length: ", fft.Len())

		// fft.Coefficients(complexFrames[i], goingHam[i])
	}

	// https://en.wikipedia.org/wiki/Triangular_function
	// Convolution: Operation that combines two signals to produce a third signal. - this should include/be the overlap-add step.

	// https://pkg.go.dev/github.com/brettbuddin/fourier@v0.1.1#section-readme
	// This is a possible library?
	// also this library does FFT and they don't use make[][], just a single make[]. So I'm not sure if I'm doing it correctly.

	// END FFT
	// Start Power Spectrum - Triangular filtering Pre-Step

	// I'm squaring the magnitude (so abs value) of the complex numbers, But im extracting the real numbers and discarding the imaginary numbers.
	// Import the math library

	powerSpectrum := make([][]float64, len(fftCoefficients))
	for i := 0; i < len(fftCoefficients); i++ {
		powerSpectrum[i] = make([]float64, len(fftCoefficients[i]))
		for j := 0; j < len(fftCoefficients[i]); j++ {
			magnitude := math.Sqrt(real(fftCoefficients[i][j])*real(fftCoefficients[i][j]) + imag(fftCoefficients[i][j])*imag(fftCoefficients[i][j]))
			powerSpectrum[i][j] = magnitude * magnitude
		}
	}
	//fmt.Println("Power Spectrum: ", powerSpectrum)

	// End Power Spectrum

	// Start Triangular Filtering - Mel Filter Banks
	// https://www.statistics.com/glossary/triangular-filter/#:~:text=As%20compared%20to%20the%20rectangular,short%20course%20Time%20Series%20Forecasting%20
	// https://en.wikipedia.org/wiki/Mel_scale
	// https://en.wikipedia.org/wiki/Periodogram

	// Take the magnitude squared of the complex Fourier coefficients - This is the power spectrum
	// Map the power spectrum onto the Mel scale using a filterbank.

	// https://en.wikipedia.org/wiki/Mel_scale
	//melScale := 2595 * math.Log10(1+(sampleRate/2)/700) // Divide by 2 for Nyquist frequency
	//filterBankSize := 26                                // Number of filters

	// NOTE! This is NOT like windowing/framing. Framinng we can think of as something we applied horizontally all along the data. The filter(s) we are applying we apply vertically.

	// So first apply the melscale to the power spectrum?
	// then find values for the filter banks?
	// then apply the filter banks to the power spectrum?

	// Map the power spectrum onto the Mel scale
	// melScale := 2595 * math.Log10(1+(each_index_of_powerSpectrum/2)/700) // Divide by 2 for Nyquist frequency

	// May want to make a new slice to have a new name for clarity
	for i := 0; i < len(powerSpectrum); i++ {
		for j := 0; j < len(powerSpectrum[i]); j++ {
			powerSpectrum[i][j] = 2595 * math.Log10(1+(float64(powerSpectrum[i][j])/2)/700) // Divide by 2 for Nyquist frequency
		}
	}

	// ---Filter Bank Process---
	// The filter banks are vertical?
	// Choose relevant filter bank values based on...?
	// Apply the filter banks to the power spectrum
	// So for the filtering, we calculate the weights based on adjacent indices?

	// For range of PowerSpectrum
	// multiply x adjacent indices by the filter bank values
	// add them together to get the weighted sum?
	// take the log of the result

	// Does every filterbank get applied across the entire spectrum? or is it dynamic, one filter per "section size"?
	// Meaning the entire set of filterbanks applied to every part of the spectrum.

	// End Triangular Filtering
	// [0, 1, 2, 3, 4, 5, 6, 7, 8, 9] - indexes
	// [1, 1, 1, 1, 1, 1, 1, 1, 1, 1] - values

	// Weighted output through summing
	// [0+12, 0+1+23, 01+2+34, 12+3+45, 23+4+56, 34+5+67, 45+6+78, 56+7+89, 67+8+9, 78+9]

	// [0.000 0.053 0.105 0.158 0.211 0.263 0.316 0.368 0.421 0.474 0.526 0.579 0.632 0.684 0.737 0.789 0.842 0.895 0.947 1.000, 0.947,]

	// Add 20 zeros to the beginning and end of the power spectrum as padding for the kernel
	paddedSpectrum := make([][]float64, len(powerSpectrum))
	for i := 0; i < len(powerSpectrum); i++ {
		temp := make([]float64, len(powerSpectrum[i])+40)
		copy(temp[20:], powerSpectrum[i])
		paddedSpectrum[i] = temp
	}

	// My understanding is that for the kernel, if it is a kernel of width 40, the actual length is 41 so that the center is 1.0?
	triangularkernel := []float64{0.05, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5, 0.55, 0.6, 0.65, 0.7, 0.75, 0.8, 0.85, 0.9, 0.95, 1.0, 0.95, 0.9, 0.85, 0.8, 0.75, 0.7, 0.65, 0.6, 0.55, 0.50, 0.45, 0.40, 0.35, 0.30, 0.25, 0.20, 0.15, 0.10, 0.05}

	fmt.Println("I made it to here")
	weightedOutput := make([][]float64, len(powerSpectrum))
	for i := 0; i < len(powerSpectrum); i++ { // Index of slices
		weightedOutput[i] = make([]float64, len(powerSpectrum[i]))
		for j := 0; j < len(powerSpectrum[i]); j++ { // Index of values
			for k := 0; k < len(triangularkernel); k++ { // Index of kernel
				weightedOutput[i][j] += paddedSpectrum[i][j+k] * triangularkernel[k]

			}
		}
	}

	// End Triangular Filtering
	// Take Log // Turn into decibels??

	for i := 0; i < len(weightedOutput); i++ {
		for j := 0; j < len(weightedOutput[i]); j++ {
			weightedOutput[i][j] = math.Log(weightedOutput[i][j])
		}
	}
	// End Take Log

	// Start DCT - Discrete Cosine Transform
	// https://pkg.go.dev/gonum.org/v1/gonum@v0.14.0/dsp/fourier#DCT
	dctCoefficients := make([][]float64, len(weightedOutput))

	for i := 0; i < len(weightedOutput); i++ {
		dct := fourier.NewDCT(len(weightedOutput[i]))
		dct.Transform(dctCoefficients[i], weightedOutput[i])
		//dctCoeffcients[i] = dct.Transform(nil, weightedOutput[i])

	}
	// End DCT

}

// Important notes to revisit
// Dynamic vs Static Framing
// Discard half of the FFT? Only compute the first half of the FFT? Only the first half contains unique information?
// filterbank/more filters? Frequency bins is huh?
// Do i multiply the Log of the energies by 10? or is it fine?
