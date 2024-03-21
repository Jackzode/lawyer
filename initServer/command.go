package initServer

var (
	// dataDirPath save all answer application data in this directory. like config file, upload file...
	dataDirPath = "./jack-test-data"
	// dumpDataPath dump data path
	dumpDataPath = "./jack-test-dump"
	// plugins needed to build in answer application
	buildWithPlugins []string
	// build output path
	buildOutput = ""
	// This config is used to upgrade the database from a specific version manually.
	// If you want to upgrade the database to version 1.1.0, you can use `answer upgrade -f v1.1.0`.
	upgradeVersion = ""
	// The fields that need to be set to the default value
	configFields []string
)

/*
func runCmd() {
	cli.FormatAllPath(dataDirPath)
	fmt.Println("config file path: ", cli.GetConfigFilePath())
	fmt.Println("Answer is starting..........................")
}



// 这个就是初始化的入口，完成初始化后程序退出
func initCmd() {
	cli.InstallAllInitialEnvironment(dataDirPath)
	configFileExist := cli.CheckConfigFile(cli.GetConfigFilePath())
	//如果配置文件存在，则测试一下db是否通，通则直接返回，认为初始化成功，不通则继续进行初始化
	if configFileExist {
		fmt.Println("config file exists, try to read the config...")
		c, err := conf.ReadConfig(cli.GetConfigFilePath())
		if err != nil {
			fmt.Println("read config failed: ", err.Error())
			return
		}

		fmt.Println("config file read successfully, try to connect database...")
		if cli.CheckDBTableExist(c.Data.Database) {
			fmt.Println("connect to database successfully and table already exists, do nothing.")
			return
		}
	}
	// start installation server to install
	///conf/config.yaml
	install.Run(cli.GetConfigFilePath())
}

func upgradeCmd() {
	log.SetLogger(log.NewStdLogger(os.Stdout))
	cli.FormatAllPath(dataDirPath)
	cli.InstallI18nBundle(true)
	c, err := conf.ReadConfig(cli.GetConfigFilePath())
	if err != nil {
		fmt.Println("read config failed: ", err.Error())
		return
	}
	if err = migrations.Migrate(c.Debug, c.Data.Database, c.Data.Cache, upgradeVersion); err != nil {
		fmt.Println("migrate failed: ", err.Error())
		return
	}
	fmt.Println("upgrade done")
}

func dumpCmd() {
	fmt.Println("Answer is backing up data")
	cli.FormatAllPath(dataDirPath)
	c, err := conf.ReadConfig(cli.GetConfigFilePath())
	if err != nil {
		fmt.Println("read config failed: ", err.Error())
		return
	}
	err = cli.DumpAllData(c.Data.Database, dumpDataPath)
	if err != nil {
		fmt.Println("dump failed: ", err.Error())
		return
	}
	fmt.Println("Answer backed up the data successfully.")
}


func checkCmd() {
	cli.FormatAllPath(dataDirPath)
	fmt.Println("Start checking the required environment...")
	if cli.CheckConfigFile(cli.GetConfigFilePath()) {
		fmt.Println("config file exists [✔]")
	} else {
		fmt.Println("config file not exists [x]")
	}

	if cli.CheckUploadDir() {
		fmt.Println("upload directory exists [✔]")
	} else {
		fmt.Println("upload directory not exists [x]")
	}

	c, err := conf.ReadConfig(cli.GetConfigFilePath())
	if err != nil {
		fmt.Println("read config failed: ", err.Error())
		return
	}

	if cli.CheckDBConnection(c.Data.Database) {
		fmt.Println("db connection successfully [✔]")
	} else {
		fmt.Println("db connection failed [x]")
	}
	fmt.Println("check environment all done")

}

func buildCmd() {
	fmt.Printf("try to build a new answer with plugins:\n%s\n", strings.Join(buildWithPlugins, "\n"))
	err := cli.BuildNewAnswer(buildOutput, buildWithPlugins, cli.OriginalAnswerInfo{
		Version:  Version,
		Revision: Revision,
		Time:     Time,
	})
	if err != nil {
		fmt.Printf("build failed %v", err)
	} else {
		fmt.Printf("build new answer successfully %s\n", buildOutput)
	}
}

func pluginCmd() {
	_ = plugin.CallBase(func(base plugin.Base) error {
		info := base.Info()
		fmt.Printf("%s[%s] made by %s\n", info.SlugName, info.Version, info.Author)
		return nil
	})
}


func configCmd() {
	cli.FormatAllPath(dataDirPath)

	c, err := conf.ReadConfig(cli.GetConfigFilePath())
	if err != nil {
		fmt.Println("read config failed: ", err.Error())
		return
	}
	field := &cli.ConfigField{}
	for _, f := range configFields {
		switch f {
		case "allow_password_login":
			//field.AllowPasswordLogin = true
		default:
			//fmt.Printf("field %s not support\n", f)
		}
	}
	//err = cli.SetDefaultConfig(c.Data.Database, c.Data.Cache, field)
	//if err != nil {
		fmt.Println("set default config failed: ", err.Error())
	} else {
		fmt.Println("set default config successfully")
	}
}
*/
