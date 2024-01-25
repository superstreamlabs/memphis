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

import React, { useContext, useEffect, useState } from 'react';

import OverflowTip from 'components/tooltip/overflowtip';
import Reducer from 'hooks/reducer';
import { StationStoreContext } from 'domain/stationOverview';
import { parsingDate } from 'services/valueConvertor';

const GenericList = (props) => {
    const [stationState] = useContext(StationStoreContext);

    const { columns, tab } = props;
    const [rowsData, setRowsData] = useState([]);

    useEffect(() => {
        if (tab === 0) {
            setRowsData(stationState?.stationSocketData?.audit_logs);
        }
    }, [stationState]);

    return (
        <div className="generic-list-wrapper">
            <div className="list">
                <div className="coulmns-table">
                    {columns?.map((column, index) => {
                        return (
                            <span key={index} style={{ width: column.width }}>
                                {column.title}
                            </span>
                        );
                    })}
                </div>
                <div className="rows-wrapper">
                    {rowsData?.map((row, index) => {
                        return (
                            <div className="pubSub-row" key={index}>
                                <OverflowTip text={row?.message || row?.produced_by} width={'300px'}>
                                    {row?.message || row?.produced_by}
                                </OverflowTip>
                                <OverflowTip text={row?.created_by_username || row?.consumer} width={'200px'}>
                                    {row?.created_by_username || row?.consumer}
                                </OverflowTip>
                                <OverflowTip text={parsingDate(row?.created_at)} width={'200px'}>
                                    {parsingDate(row?.created_at)}
                                </OverflowTip>
                            </div>
                        );
                    })}
                </div>
            </div>
        </div>
    );
};

export default GenericList;
