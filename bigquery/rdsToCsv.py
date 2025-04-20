import sys
from awsglue.context import GlueContext
from awsglue.utils import getResolvedOptions
from awsglue.job import Job
from pyspark.context import SparkContext

# 引数取得
args = getResolvedOptions(sys.argv, ['JOB_NAME', 'tables', 'database', 's3_output_path'])

table_names = args['tables'].split(',')
database = args['database']
s3_output_path = args['s3_output_path']

# Glue/Sparkコンテキストの初期化
sc = SparkContext()
glueContext = GlueContext(sc)
spark = glueContext.spark_session

job = Job(glueContext)
job.init(args['JOB_NAME'], args)

# 各テーブルを処理
for table_name in table_names:
    print(f"Processing table: {table_name}")
    
    # Glue CatalogからDynamicFrame取得
    dyf = glueContext.create_dynamic_frame.from_catalog(
        database=database,
        table_name=table_name
    )
    
    # 出力先パス
    output_path = f"{s3_output_path.rstrip('/')}/{table_name}/"
    
    # CSVでS3に書き出し
    glueContext.write_dynamic_frame.from_options(
        frame=dyf,
        connection_type="s3",
        connection_options={"path": output_path},
        format="csv"
    )

job.commit()


# --extra-py-files s3://my-script/table_list.json --arguments --tables table1,table2,table3
# --database mydb
# --s3_output_path s3://your-bucket/output/
