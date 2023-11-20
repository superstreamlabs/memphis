// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useEffect, useState } from 'react';
import { ApiEndpoints } from '../../../../../const/apiEndpoints';
import { httpRequest } from '../../../../../services/http';
import CustomTabs from '../../../../../components/Tabs';
import FunctionLogs from '../functionLogs';
import { ReactComponent as CloseIcon } from '../../../../../assets/images/close.svg';
import { ReactComponent as MetricsIcon } from '../../../../../assets/images/metricsIcon.svg';
import { ReactComponent as MetricsClockIcon } from '../../../../../assets/images/metricsClockIcon.svg';
import { ReactComponent as MetricsErrorIcon } from '../../../../../assets/images/metricsErrorIcon.svg';

const tabValuesList = ['Information', 'Logs', 'Dead-letter'];

const FunctionData = ({ functionDetails }) => {
    const [tabValue, setTabValue] = useState('Information');

    return (
        <div className="function-data-container">
            <CustomTabs tabs={tabValuesList} size={'small'} tabValue={tabValue} onChange={(tabValue) => setTabValue(tabValue)} />
            {tabValue === tabValuesList[0] && (
                <div className="metrics-wrapper">
                    <div className="metrics">
                        <div className="metrics-img">
                            <MetricsIcon />
                        </div>
                        <div className="metrics-body">
                            <div className="metrics-body-title">Total invocations</div>
                            <div className="metrics-body-subtitle">{functionDetails?.metrics?.total_invocations?.toLocaleString() || 0}</div>
                        </div>
                    </div>
                    <div className="metrics-divider"></div>
                    <div className="metrics">
                        <div className="metrics-img">
                            <MetricsClockIcon />
                        </div>
                        <div className="metrics-body">
                            <div className="metrics-body-title">Av. Processing time</div>
                            <div className="metrics-body-subtitle">
                                {functionDetails?.metrics?.average_processing_time}
                                <span>/sec</span>
                            </div>
                        </div>
                    </div>
                    <div className="metrics-divider"></div>
                    <div className="metrics">
                        <div className="metrics-img">
                            <MetricsErrorIcon />
                        </div>
                        <div className="metrics-body">
                            <div className="metrics-body-title">Error rate</div>
                            <div className="metrics-body-subtitle">{functionDetails?.metrics?.error_rate}%</div>
                        </div>
                    </div>
                </div>
            )}
            {tabValue === tabValuesList[1] && <FunctionLogs functionId={functionDetails?.function?.id} />}
        </div>
    );
};

export default FunctionData;
