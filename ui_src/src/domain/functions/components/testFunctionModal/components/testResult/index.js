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
import { ReactComponent as StatusIcon } from '../../../../../../assets/images/statusIcon.svg';
import Tag from '../../../../../../components/tag';
import RadioButton from '../../../../../../components/radioButton';
import Spinner from '../../../../../../components/spinner';
import Copy from '../../../../../../components/copy';

import { ColorPalette } from '../../../../../../const/globalConst';

const options = [
    {
        label: 'All',
        value: 'All'
    },
    {
        label: 'Success',
        value: 'Success'
    },
    {
        label: 'Failure',
        value: 'Failure'
    },
    {
        label: 'Logs',
        value: 'Logs'
    }
];

const TestResult = ({ testResultData, loading }) => {
    const [responseTab, setResponseTab] = useState('All');
    const [testResult, setTestResult] = useState(null);

    useEffect(() => {
        setTestResult(testResultData);
    }, [testResultData]);

    const getCopyData = () => {
        if (responseTab === 'All') {
            return JSON.stringify(testResult, null, 2);
        } else if (responseTab === 'Success') {
            return JSON.stringify(testResult?.messages, null, 2);
        } else if (responseTab === 'Failure') {
            return JSON.stringify(testResult?.failed_messages, null, 2);
        } else if (responseTab === 'Logs') {
            return testResult?.logs;
        } else {
            return testResult?.error;
        }
    };

    const getSuccessMessages = () => {
        return (
            testResult?.messages && (
                <>
                    <p className="title">Success</p>
                    {testResult?.messages?.map((message, index) => {
                        return <p key={`message-${index}`}>{message?.payload}</p>;
                    })}
                </>
            )
        );
    };

    const getFailedMessages = () => {
        return (
            testResult?.failed_messages && (
                <>
                    <p className="title">Failure</p>
                    {testResult?.failed_messages?.map((message, index) => {
                        return <p key={`failed_messages-${index}`}>{message?.payload}</p>;
                    })}
                </>
            )
        );
    };

    const getLogs = () => {
        return (
            testResult?.logs && (
                <>
                    <p className="title">Logs</p>
                    <p>{testResult?.logs}</p>
                </>
            )
        );
    };

    const getErrors = () => {
        return (
            testResult?.error && (
                <>
                    <p className="title">Error</p>
                    <p>{testResult?.error}</p>
                </>
            )
        );
    };

    return (
        <div className="result-wrapper">
            <RadioButton
                vertical={false}
                height="25px"
                fontFamily="InterSemiBold"
                options={!testResult || testResult?.error === '' ? options : [{ label: 'Error', value: 'Error' }]}
                radioStyle="radiobtn-capitalize"
                radioValue={responseTab}
                onChange={(e) => setResponseTab(e.target.value)}
            />
            <div className="result-container">
                <div className="header">
                    <p className="title">EXECUTION RESULT</p>

                    <div className="rightSide">
                        <StatusIcon />

                        {testResult && (
                            <>
                                <span className="status">Status:</span>
                                <Tag
                                    tag={{
                                        name: testResult?.failed_messages ? 'Failed' : 'Successful',
                                        color: testResult?.failed_messages ? ColorPalette[7] : ColorPalette[9]
                                    }}
                                    editable={false}
                                    rounded={false}
                                />
                            </>
                        )}
                    </div>
                </div>
                {!loading && (
                    <div className="result">
                        <div className="copy-section">
                            <Copy data={getCopyData()} />
                        </div>
                        {responseTab === 'Error' && getErrors()}
                        {responseTab === 'All' && (
                            <span>
                                {getSuccessMessages()}
                                {getFailedMessages()}
                                {getLogs()}
                            </span>
                        )}
                        {responseTab === 'Success' && getSuccessMessages()}
                        {responseTab === 'Failure' && getFailedMessages()}
                        {responseTab === 'Logs' && getLogs()}
                    </div>
                )}
                {loading && (
                    <div className="loader-wrapper">
                        <Spinner />
                    </div>
                )}
            </div>
        </div>
    );
};

export default TestResult;
