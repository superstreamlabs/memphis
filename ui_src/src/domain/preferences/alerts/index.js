// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import React, { useState } from 'react';

import Switcher from '../../../components/switcher';

const Alerts = () => {
    const [errorsAlert, setErrorsAlert] = useState(false);
    const [schemaAlert, setSchemaAlert] = useState(false);
    return (
        <div className="alerts-integrations-container">
            <h3 className="title">We will keep an eye on your data streams and alert you if anything went wrong according to the following triggers:</h3>
            <div>
                <div className="alert-integration-type">
                    <label className="alert-label-bold">Errors</label>
                    <Switcher onChange={() => setErrorsAlert(!errorsAlert)} checked={errorsAlert} checkedChildren="on" unCheckedChildren="off" />
                </div>
                <div className="alert-integration-type">
                    <label className="alert-label-bold">Schema has changed</label>
                    <Switcher onChange={() => setSchemaAlert(!schemaAlert)} checked={schemaAlert} checkedChildren="on" unCheckedChildren="off" />
                </div>
            </div>
        </div>
    );
};

export default Alerts;
