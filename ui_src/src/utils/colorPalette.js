const colors = {
    light: {
        primaryBlue: "#5A4FE5",
        sunglow: "#FFC633",
        greenishCyan: "#61DFC6",
        sandyBeach: "#FDEDC2",
        success: "#34C759",
        warning: "#FF9F38",
        failed: "#FF3B30",
        mainText: "#1D1D1D",
        greyLight: "#979797",
        greyDark: "#484848",
        white: "#FFFFFF",
        mainBG: "#F9FAFB",
        card: "#FFFFFF",
        lineColor: "#E1E1E1"
    },
    dark: {
        primaryBlue: "#5A4FE5",
        sunglow: "#FFC633",
        greenishCyan: "#61DFC6",
        sandyBeach: "#FDEDC2",
        success: "#34C759",
        warning: "#FF9F38",
        failed: "#FF3B30",
        mainText: "#FFFFFF",
        greyLight: "#979797",
        greyDark: "#484848",
        white: "#1D1D1D",
        mainBG: "#151719",
        card: "#212325",
        card2: "#2E3032",
        lineColor: "#323436"
    }
};

function designSystemColor(darkMode) {
    return darkMode ? colors.dark : colors.light;
}

module.exports = designSystemColor;
