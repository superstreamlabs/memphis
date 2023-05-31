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

import React from 'react';
import Button from '../../../../../../components/button';
import FileDownload from '../../../../../../assets/images/setting/file_download.svg';
import Filter from '../../../../../../components/filter';

function InvoiceHistory() {
    return (
        <div className="invoice-history-container">
            <div className="invoice-history-header">
                <div>
                    <p className="invoice-history-title">Invoice History</p>
                    <p className="invoice-history-description">Contrary to popular belief, Lorem Ipsum</p>
                </div>
                <div className="header-filter">
                    <Filter filterComponent="invoices" height="34px" />
                    <Button
                        className="modal-btn"
                        width="150px"
                        height="32px"
                        placeholder={
                            <div>
                                <img className="download-img" src={FileDownload} alt="Generate Report" />
                                Generate Report
                            </div>
                        }
                        disabled={false}
                        colorType="navy"
                        radiusType="circle"
                        border="gray"
                        backgroundColorType={'white'}
                        fontSize="12px"
                        fontWeight="600"
                        isLoading={false}
                        onClick={() => {
                            console.log('hi');
                        }}
                    />
                </div>
            </div>
        </div>
    );
}

export default InvoiceHistory;
