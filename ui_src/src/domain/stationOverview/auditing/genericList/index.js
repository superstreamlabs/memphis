// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';

import React, { useContext, useEffect, useState } from 'react';

import OverflowTip from '../../../../components/tooltip/overflowtip';
import Reducer from '../../hooks/reducer';
import { StationStoreContext } from '../..';
import { parsingDate } from '../../../../services/valueConvertor';

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
                                <OverflowTip text={row?.user_type || row?.consumer} width={'200px'}>
                                    {row?.user_type || row?.consumer}
                                </OverflowTip>
                                <OverflowTip text={parsingDate(row?.creation_date)} width={'200px'}>
                                    {parsingDate(row?.creation_date)}
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
