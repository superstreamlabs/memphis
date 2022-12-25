export const Sleep = (sec) => {
    return new Promise((resolve, reject) => {
        setTimeout(() => resolve(), sec * 1000);
    });
};
