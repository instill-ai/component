package video

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/instill-ai/component/base"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"google.golang.org/protobuf/types/known/structpb"
)

type SubsampleVideoInput struct {
	Video     Video  `json:"video"`
	Fps       int    `json:"fps"`
	StartTime string `json:"start-time"`
	Duration  string `json:"duration"`
}

type SubsampleVideoOutput struct {
	Video Video `json:"video"`
}

type SubsampleVideoFramesInput struct {
	Video     Video  `json:"video"`
	Fps       int    `json:"fps"`
	StartTime string `json:"start-time"`
	Duration  string `json:"duration"`
}

type SubsampleVideoFramesOutput struct {
	Frames []Frame `json:"frames"`
}

// Base64 encoded video
type Video string

// Base64 encoded frame
type Frame string

func subsampleVideo(input *structpb.Struct) (*structpb.Struct, error) {

	inputStruct := SubsampleVideoInput{}

	err := base.ConvertFromStructpb(input, &inputStruct)
	if err != nil {
		return nil, fmt.Errorf("error converting input to struct: %v", err)
	}

	base64Video := string(inputStruct.Video)

	videoBytes, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(base64Video))

	if err != nil {
		return nil, fmt.Errorf("error in decoding for inner: %s", err)
	}

	videoPrefix := strings.Split(base64Video, ",")[0]

	// TODO: chuang8511 map the file extension to the correct format
	// using mp4 first because it has smaller size.
	tempFileIn := "temp.mp4"

	err = os.WriteFile(tempFileIn, videoBytes, 0644)
	if err != nil {
		return nil, fmt.Errorf("error in writing file: %s", err)
	}

	split := ffmpeg.Input(tempFileIn)

	tempFileOut := "temp_out.mp4"

	err = split.Output(tempFileOut, getKwArgs(inputStruct)).Run()

	if err != nil {
		return nil, fmt.Errorf("error in running ffmpeg: %s", err)
	}

	byOut, _ := os.ReadFile(tempFileOut)
	base64Subsample := videoPrefix + "," + base64.StdEncoding.EncodeToString(byOut)

	os.Remove(tempFileIn)
	os.Remove(tempFileOut)

	output := SubsampleVideoOutput{
		Video: Video(base64Subsample),
	}

	return base.ConvertToStructpb(output)
}

func getKwArgs(inputStruct SubsampleVideoInput) ffmpeg.KwArgs {
	kwArgs := ffmpeg.KwArgs{"r": inputStruct.Fps}
	if inputStruct.StartTime != "" {
		kwArgs["ss"] = inputStruct.StartTime
	}
	if inputStruct.Duration != "" {
		kwArgs["t"] = inputStruct.Duration
	}
	return kwArgs
}

func subsampleVideoFrames(input *structpb.Struct) (*structpb.Struct, error) {
	inputStruct := SubsampleVideoFramesInput{}

	err := base.ConvertFromStructpb(input, &inputStruct)
	if err != nil {
		return nil, fmt.Errorf("error converting input to struct: %v", err)
	}

	base64Video := string(inputStruct.Video)

	videoBytes, err := base64.StdEncoding.DecodeString(base.TrimBase64Mime(base64Video))

	if err != nil {
		return nil, fmt.Errorf("error in decoding for inner: %s", err)
	}

	tempFileIn := "temp.mp4"
	defer os.Remove(tempFileIn)

	err = os.WriteFile(tempFileIn, videoBytes, 0644)
	if err != nil {
		return nil, fmt.Errorf("error in writing file: %s", err)
	}

	// TODO: chuang8511 confirm how to handle the output pattern
	// Now, it only contains 4 digits, which means it can only handle 9999 frames
	outputPattern := "frame_%04d.jpeg"

	err = ffmpeg.Input(tempFileIn).
		Output(outputPattern,
			getFramesKwArgs(inputStruct),
		).
		Run()

	if err != nil {
		return nil, fmt.Errorf("error in running ffmpeg: %s", err)
	}

	files, err := filepath.Glob("frame_*.jpeg")
	if err != nil {
		return nil, fmt.Errorf("error listing frames: %s", err)
	}

	sort.Strings(files)
	jpegPrefix := "data:image/jpeg;base64,"
	var frames []Frame
	for _, file := range files {

		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", file, err)
		}

		encoded := base64.StdEncoding.EncodeToString(data)

		frames = append(frames, Frame(jpegPrefix+encoded))

		err = os.Remove(file)
		if err != nil {
			return nil, fmt.Errorf("error removing file %s: %v", file, err)
		}
	}

	output := SubsampleVideoFramesOutput{
		Frames: frames,
	}

	return base.ConvertToStructpb(output)
}

func getFramesKwArgs(inputStruct SubsampleVideoFramesInput) ffmpeg.KwArgs {
	kwArgs := ffmpeg.KwArgs{"vf": "fps=" + fmt.Sprintf("%d", inputStruct.Fps)}
	if inputStruct.StartTime != "" {
		kwArgs["ss"] = inputStruct.StartTime
	}
	if inputStruct.Duration != "" {
		kwArgs["t"] = inputStruct.Duration
	}
	return kwArgs
}
