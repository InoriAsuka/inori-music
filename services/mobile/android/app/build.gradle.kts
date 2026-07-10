import java.util.Properties

plugins {
    id("com.android.application")
    // The Flutter Gradle Plugin must be applied after the Android and Kotlin Gradle plugins.
    id("dev.flutter.flutter-gradle-plugin")
}

@Suppress("DEPRECATION")
android {
    namespace = "com.inori.music.inori_music"
    compileSdk = flutter.compileSdkVersion
    ndkVersion = flutter.ndkVersion

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    defaultConfig {
        applicationId = "com.inori.music.inori_music"
        minSdk = flutter.minSdkVersion
        targetSdk = flutter.targetSdkVersion
        versionCode = flutter.versionCode
        versionName = flutter.versionName
    }

    // Release signing: reads keystore path/credentials from environment or
    // local key.properties file (gitignored). Falls back to debug signing so
    // `flutter build apk --release` works without credentials (e.g. in CI
    // smoke tests).
    val keystoreFile = System.getenv("ANDROID_KEYSTORE_PATH")?.let { file(it) }
        ?: rootProject.file("key.properties").takeIf { it.exists() }?.let { props ->
            Properties().also { it.load(props.inputStream()) }
            null  // handled below
        }

    val keyProps = rootProject.file("key.properties").takeIf { it.exists() }?.let {
        Properties().also { p -> p.load(it.inputStream()) }
    }

    signingConfigs {
        if (System.getenv("ANDROID_KEYSTORE_PATH") != null) {
            create("release") {
                storeFile = file(System.getenv("ANDROID_KEYSTORE_PATH")!!)
                storePassword = System.getenv("ANDROID_STORE_PASSWORD") ?: ""
                keyAlias = System.getenv("ANDROID_KEY_ALIAS") ?: ""
                keyPassword = System.getenv("ANDROID_KEY_PASSWORD") ?: ""
            }
        } else if (keyProps != null) {
            create("release") {
                storeFile = file(keyProps.getProperty("storeFile", ""))
                storePassword = keyProps.getProperty("storePassword", "")
                keyAlias = keyProps.getProperty("keyAlias", "")
                keyPassword = keyProps.getProperty("keyPassword", "")
            }
        }
    }

    buildTypes {
        release {
            signingConfig = if (signingConfigs.names.contains("release"))
                signingConfigs.getByName("release")
            else
                signingConfigs.getByName("debug")
        }
    }
}

kotlin {
    compilerOptions {
        jvmTarget = org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17
    }
}

flutter {
    source = "../.."
}
