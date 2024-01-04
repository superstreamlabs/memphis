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
import Modal from 'components/modal';
import { IoRocket } from 'react-icons/io5';
import Copy from 'components/copy';
import CustomTabs from 'components/Tabs';
import { githubUrls } from 'const/globalConst';
import { SiLinux, SiApple, SiWindows11 } from 'react-icons/si';
import { LOCAL_STORAGE_BROKER_HOST, LOCAL_STORAGE_ENV, LOCAL_STORAGE_ACCOUNT_ID } from 'const/localStorageConsts';

let write =
    'mem bench producer --message-size 128 --count 1000 --concurrency 1 --host <host> --account-id <account-id(not needed for open-source)> --user <client type user> --password <password>';
let read =
    'mem bench consumer --message-size 128 --count 1000 --concurrency 1 --batch-size 50 --host <host> --account-id <account-id(not needed for open-source)> --user <client type user> --password <password>';

const RunBenchmarkModal = ({ open, clickOutside }) => {
    const [tabValue, setTabValue] = useState('Windows');
    const [writeLink, setWriteLink] = useState(null);
    const [readLink, setReadLink] = useState(null);

    useEffect(() => {
        let host =
            localStorage.getItem(LOCAL_STORAGE_ENV) === 'docker'
                ? 'localhost'
                : localStorage.getItem(LOCAL_STORAGE_BROKER_HOST)
                ? localStorage.getItem(LOCAL_STORAGE_BROKER_HOST)
                : 'memphis.memphis.svc.cluster.local';
        write = write.replace('<host>', host);
        write = write.replace('<account-id(not needed for open-source)>', parseInt(localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID)));
        read = write.replace('<host>', host);
        read = write.replace('<account-id(not needed for open-source)>', parseInt(localStorage.getItem(LOCAL_STORAGE_ACCOUNT_ID)));
        setWriteLink(write);
        setReadLink(read);
    }, []);

    return (
        <Modal
            header={
                <div className="modal-header connector-modal-header">
                    <div className="header-img-container">
                        <IoRocket className="headerImage" alt="stationImg" style={{ color: '#6557FF' }} />
                    </div>
                    <div className="connector-modal-title">
                        <div className="modal-title">Run a benchmark</div>
                    </div>
                </div>
            }
            className={'modal-wrapper'}
            width="550px"
            clickOutside={clickOutside}
            open={open}
            displayButtons={true}
            rBtnText={'Close'}
            rBtnClick={clickOutside}
        >
            <div className="benchmark-wrapper">
                <CustomTabs
                    tabs={['Windows', 'Mac', 'Linux RPM', 'Linux APK']}
                    icons={[<SiWindows11 />, <SiApple />, <SiLinux />, <SiLinux />]}
                    size={'small'}
                    tabValue={tabValue}
                    onChange={(tabValue) => setTabValue(tabValue)}
                />

                <>
                    <p className="action">
                        <label>Step 1:</label> Install Memphis CLI
                    </p>
                    <div className="url-wrapper">
                        <p className="url-text"> {githubUrls['cli'][tabValue]}</p>
                        <div className="icon">
                            <Copy width="18" data={githubUrls['cli'][tabValue]} />
                        </div>
                    </div>
                    <p className="action">
                        <label>Step 2:</label> For inspecting write latency, run the following
                    </p>
                    <div className="url-wrapper">
                        <p className="url-text">{writeLink}</p>
                        <div className="icon">
                            <Copy width="18" data={writeLink || write} />
                        </div>
                    </div>
                    <p className="action">
                        <label>Step 3:</label> For inspecting read latency, run the following
                    </p>
                    <div className="url-wrapper">
                        <p className="url-text">{readLink || read}</p>
                        <div className="icon">
                            <Copy width="18" data={readLink || read} />
                        </div>
                    </div>
                </>

                <span>
                    <p className="subtitle">Considerations to keep in mind:</p>
                    <p className="subtitle">1. The latency and throughput largely depend on the internet connection.</p>
                    <p className="subtitle">
                        2.These figures are preliminary and subject to improvement if necessary. Consult with an engineer for further optimization.
                    </p>
                </span>
            </div>
        </Modal>
    );
};

export default RunBenchmarkModal;
