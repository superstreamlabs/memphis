export const entitlementChecker = (object, entitlement) => {
    const entitlements = object?.userData?.entitlements?.entitelments || object?.userData?.entitlements?.entitlements || object?.userData?.entitlements;

    return entitlements ? entitlements.hasOwnProperty(entitlement) : false;
};