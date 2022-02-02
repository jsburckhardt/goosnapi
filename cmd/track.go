package cmd

import (
	"log"
	"time"

	"goosnapi/pkg/openskyapi"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// trackCmd represents the track command
var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "Prints flight records based on a given region",
	Long:  "Prints flight records based on a given region",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("track called")
		coverArea := openskyapi.BoundBox{
			MinLatitude:  viper.GetFloat64("opensky.minlatitude"),
			MaxLatitude:  viper.GetFloat64("opensky.maxlatitude"),
			MinLongitude: viper.GetFloat64("opensky.minlongitude"),
			MaxLongitude: viper.GetFloat64("opensky.maxlongitude"),
		}
		err := coverArea.Validate()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Monitoring area: %+v\n", coverArea)
		for {
			rawResponse, err := openskyapi.GetFlightsInArea(coverArea)
			if err != nil {
				log.Printf("error retrieving flight data. error: %+v\n", err)
				continue
			}
			structuredResponse := openskyapi.StructureRawResponse(rawResponse)
			err = structuredResponse.Print()
			if err != nil {
				log.Printf("error printing flight data. error: %+v\n", err)
				continue
			}
			time.Sleep(time.Duration(viper.GetInt("frequency")) * time.Second)
		}
	},
}

func init() {
	var err error
	rootCmd.AddCommand(trackCmd)
	trackCmd.Flags().Float64("minlatitude", -40.110403, "defines he minimum latitude for the monitored area")
	trackCmd.Flags().Float64("maxlatitude", -24.267845, "defines the maximum latitude for the monitored area")
	trackCmd.Flags().Float64("minlongitude", 139.147805, "defines the minimum longitude for the monitored area")
	trackCmd.Flags().Float64("maxlongitude", 154.590532, "defines the maximum longitude for the monitored area")
	trackCmd.Flags().Int("frequency", 5, "The frequency to retrieve flight data. In seconds")
	err = viper.BindPFlag("opensky.minlatitude", trackCmd.Flags().Lookup("minlatitude"))
	cobra.CheckErr(err)
	err = viper.BindPFlag("opensky.maxlatitude", trackCmd.Flags().Lookup("maxlatitude"))
	cobra.CheckErr(err)
	err = viper.BindPFlag("opensky.minlongitude", trackCmd.Flags().Lookup("minlongitude"))
	cobra.CheckErr(err)
	err = viper.BindPFlag("opensky.maxlongitude", trackCmd.Flags().Lookup("maxlongitude"))
	cobra.CheckErr(err)
	err = viper.BindPFlag("frequency", trackCmd.Flags().Lookup("frequency"))
	cobra.CheckErr(err)
}
